// Package main demonstrates a multi-chat message broadcasting system in GoTG.
//
// The broadcaster sends a text message to a list of chats, handling:
//   - Peer resolution failures (skipped with logged error)
//   - FLOOD_WAIT throttling (via gotd/contrib floodwait middleware)
//   - Per-message retry with exponential backoff
//   - Progress tracking (succeeded / failed / skipped)
//   - Graceful cancellation via context
//
// Architecture:
//
//	The broadcast runs as a bounded worker pool. Each worker picks a chat ID
//	from a channel, resolves it to an InputPeer via GoTG's 3-tier lookup
//	(cache → DB → RPC), sends the message, and retries on transient errors.
//	The floodwait middleware at the transport layer handles FLOOD_WAIT_X
//	pauses automatically, so the application-level retry only needs to handle
//	other transient failures (network timeouts, etc.).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"golang.org/x/time/rate"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/storage"
)

func main() {
	apiID := 123456            // from https://my.telegram.org/apps
	apiHash := "API_HASH_HERE"
	botToken := "BOT_TOKEN_HERE"

	// Flood waiter — automatically pauses when Telegram returns FLOOD_WAIT_X.
	waiter := floodwait.NewWaiter().WithCallback(func(_ context.Context, wait floodwait.FloodWait) {
		log.Printf("FLOOD: waiting %s", wait.Duration)
	})

	// Rate limiter — cap at 20 messages per second to stay under Telegram's
	// per-second limits. Adjust based on your account type.
	limiter := ratelimit.New(rate.Every(50*time.Millisecond), 20)

	client, err := gotg.NewClient(
		apiID, apiHash,
		gotg.AsBot(botToken),
		&gotg.ClientOpts{
			Session:          session.SimpleSession(),
			PeersFromDialogs: true,
			Middlewares:      []telegram.Middleware{waiter, limiter},
			RunMiddleware: func(
				origRun func(ctx context.Context, f func(ctx context.Context) error) error,
				ctx context.Context,
				f func(ctx context.Context) error,
			) error {
				return origRun(ctx, func(ctx context.Context) error {
					return waiter.Run(ctx, f)
				})
			},
		},
	)
	if err != nil {
		log.Fatalf("client: %v", err)
	}

	dp := client.Dispatcher
	dp.AddHandler(handlers.OnCommand("broadcast", broadcastCmd(client)))

	log.Printf("Bot @%s started — send /broadcast <text> to demo", client.Self.Username)
	log.Fatal(client.Idle())
}

// broadcastCmd returns a handler that broadcasts a message to all chats
// the bot is a member of.
func broadcastCmd(client *gotg.Client) func(u *adapter.Update) error {
	return func(u *adapter.Update) error {
		text := strings.TrimSpace(strings.TrimPrefix(u.Text(), "/broadcast"))
		if text == "" {
			_, _ = u.Reply("Usage: /broadcast <message text>")
			return dispatcher.EndGroups
		}

		// Collect chat IDs — accept from command args or from peer storage.
		var chatIDs []int64
		args := u.Args()
		if len(args) > 1 {
			// Identifiers passed as command arguments.
			for _, arg := range args[1:] {
				id, err := u.ResolvePeerToID(arg)
				if err != nil {
					log.Printf("resolve %s: %v (skipping)", arg, err)
					continue
				}
				chatIDs = append(chatIDs, id)
			}
		}
		if len(chatIDs) == 0 {
			_, _ = u.Reply("Usage: /broadcast @user1 @channel1 <text>\nNo valid peers resolved.")
			return dispatcher.EndGroups
		}

		_, _ = u.Reply(fmt.Sprintf("Starting broadcast to %d chats...", len(chatIDs)))

		// Run the broadcast asynchronously so the handler returns immediately.
		go func() {
			result := RunBroadcast(u.Ctx, chatIDs, text, BroadcastConfig{
				Workers:      5,
				MaxRetries:   3,
				InitialDelay: 2 * time.Second,
				MaxDelay:     30 * time.Second,
			})

			summary := fmt.Sprintf(
				"Broadcast complete: %d sent, %d failed, %d skipped",
				result.Sent, result.Failed, result.Skipped,
			)
			_, _ = u.SendMessage(u.ChatID(), summary)
		}()

		return dispatcher.EndGroups
	}
}

// --- Broadcast types ---

// BroadcastResult holds the aggregate outcome of a broadcast.
type BroadcastResult struct {
	Total   int
	Sent    int
	Failed  int
	Skipped int
	Errors  []BroadcastError
}

// BroadcastError records a per-chat failure.
type BroadcastError struct {
	ChatID int64
	Err    error
}

// BroadcastConfig controls broadcast behaviour.
type BroadcastConfig struct {
	Workers      int           // concurrent send goroutines (default: 3)
	MaxRetries   int           // per-chat retries on transient errors (default: 2)
	InitialDelay time.Duration // first retry backoff (default: 2s)
	MaxDelay     time.Duration // backoff cap (default: 30s)
}

// RunBroadcast sends text to every chatID in the list using a bounded worker
// pool. It returns a BroadcastResult with per-chat outcomes.
//
// The caller's Context provides the raw Telegram client and PeerStorage needed
// for peer resolution and message sending. The context's embedded Go context
// controls cancellation.
func RunBroadcast(ctx *adapter.Context, chatIDs []int64, text string, cfg BroadcastConfig) BroadcastResult {
	if cfg.Workers <= 0 {
		cfg.Workers = 3
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 2
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = 2 * time.Second
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 30 * time.Second
	}

	result := BroadcastResult{Total: len(chatIDs)}
	var resultMu sync.Mutex

	// Feed chat IDs into a channel for workers to consume.
	ch := make(chan int64, len(chatIDs))
	for _, id := range chatIDs {
		ch <- id
	}
	close(ch)

	var wg sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chatID := range ch {
				sent, err := sendWithRetry(ctx, chatID, text, cfg)
				resultMu.Lock()
				if sent {
					result.Sent++
				} else if err != nil {
					result.Failed++
					result.Errors = append(result.Errors, BroadcastError{ChatID: chatID, Err: err})
				} else {
					result.Skipped++
				}
				resultMu.Unlock()
			}
		}()
	}

	wg.Wait()
	return result
}

// sendWithRetry attempts to send a message to a single chat with exponential
// backoff on transient errors. Returns (true, nil) on success,
// (false, nil) if skipped (non-retryable), or (false, err) after exhausting
// retries.
func sendWithRetry(ctx *adapter.Context, chatID int64, text string, cfg BroadcastConfig) (bool, error) {
	bo := backoff.ExponentialBackOff{
		InitialInterval:     cfg.InitialDelay,
		MaxInterval:         cfg.MaxDelay,
		Multiplier:          2.0,
		MaxElapsedTime:      0,
		Clock:               backoff.SystemClock,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
	}
	bo.Reset()

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			next := bo.NextBackOff()
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(next):
			}
		}

		// Resolve peer and send in one shot via Context.SendMessage which
		// internally calls ResolveInputPeerByID (3-tier: cache → DB → RPC).
		_, err := ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{
			Message: text,
		})
		if err == nil {
			return true, nil
		}

		lastErr = err

		if isRetryable(err) {
			continue
		}

		// Non-retryable error — skip this chat.
		return false, nil
	}

	return false, lastErr
}

// isRetryable returns true for errors that may succeed on retry.
// Non-retryable errors include auth failures, chat not found, and permissions.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// FloodWait is handled by the transport middleware, but if it leaks
	// through, treat it as retryable.
	if tgerr.Is(err, "FLOOD_WAIT") {
		return true
	}

	// Auth / permission errors — not retryable.
	nonRetryable := []string{
		"CHAT_WRITE_FORBIDDEN",
		"USER_BANNED_IN_CHANNEL",
		"CHAT_SEND_MEDIA_FORBIDDEN",
		"USER_DEACTIVATED",
		"SESSION_REVOKED",
		"AUTH_KEY_UNREGISTERED",
		"PEER_FLOOD",
		"USER_PRIVACY_RESTRICTED",
		"PEER_ID_INVALID",
		"CHANNEL_PRIVATE",
		"CHANNEL_PUBLIC_GROUP_NA",
		"USER_NOT_PARTICIPANT",
	}
	msg := err.Error()
	for _, p := range nonRetryable {
		if strings.Contains(msg, p) {
			return false
		}
	}

	// Network errors, timeouts, etc. — retryable.
	return true
}

// collectChatIDs is a placeholder that extracts chat IDs from peer storage.
// Replace this with your own subscription list, database query, or use
// ResolvePeerToID to resolve identifiers from command arguments.
func collectChatIDs(ps *storage.PeerStorage) []int64 {
	// Production alternatives:
	//   1. Accept identifiers from command args: u.ResolvePeerToID("@channel")
	//   2. Query a subscriptions table in your database
	//   3. Use PeersFromDialogs: true and iterate all stored peers
	return nil
}
