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
//	The broadcast uses the broadcast package's Broadcaster with a bounded worker
//	pool. Each worker picks a chat ID from the pool, resolves it to an InputPeer
//	via GoTG's 3-tier lookup (cache → DB → RPC), sends the message, and retries
//	on transient errors. The floodwait middleware at the transport layer handles
//	FLOOD_WAIT_X pauses automatically, so the application-level retry only needs
//	to handle other transient failures (network timeouts, etc.).
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"golang.org/x/time/rate"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/broadcast"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/session"
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
	dp.AddHandler(handlers.OnCommand("broadcast", broadcastCmd()))

	log.Printf("Bot @%s started — send /broadcast <text> to demo", client.Self.Username)
	log.Fatal(client.Idle())
}

// broadcastCmd returns a handler that broadcasts a message to all chats
// specified in the command arguments.
func broadcastCmd() func(u *adapter.Update) error {
	return func(u *adapter.Update) error {
		text := strings.TrimSpace(strings.TrimPrefix(u.Text(), "/broadcast"))
		if text == "" {
			_, _ = u.Reply("Usage: /broadcast <message text>")
			return dispatcher.EndGroups
		}

		// Collect peer identifiers from command arguments.
		args := u.Args()
		var targets []broadcast.PeerTarget
		for _, arg := range args[1:] {
			// Try to resolve the identifier.
			id, err := u.ResolvePeerToID(arg)
			if err != nil {
				log.Printf("resolve %s: %v (skipping)", arg, err)
				continue
			}
			targets = append(targets, broadcast.PeerTarget{ChatID: id})
		}
		if len(targets) == 0 {
			_, _ = u.Reply("Usage: /broadcast @user1 @channel1 <text>\nNo valid peers resolved.")
			return dispatcher.EndGroups
		}

		_, _ = u.Reply(fmt.Sprintf("Starting broadcast to %d chats...", len(targets)))

		go func() {
			b := broadcast.New(targets, broadcast.Config{
				Workers:      5,
				MaxRetries:   3,
				InitialDelay: 2 * time.Second,
				MaxDelay:     30 * time.Second,
			})

			makeSend := broadcast.MakeTextSendFunc(u.Ctx, broadcast.TextSendConfig{
				Text: text,
			})
			result := b.Run(u.Ctx, makeSend)

			summary := fmt.Sprintf(
				"Broadcast complete: %d sent, %d failed, %d skipped",
				result.Sent(), result.Failed(), result.Skipped(),
			)
			_, _ = u.SendMessage(u.ChatID(), summary)
		}()

		return dispatcher.EndGroups
	}
}
