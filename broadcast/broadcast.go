package broadcast

import (
	"context"
	"sync"
	"time"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
)

// SendFunc is a function that attempts to send a message.
// The attempt parameter is 0-indexed (0 = first attempt).
// It returns an error if the send fails.
type SendFunc func(attempt int) error

// Config controls broadcast behaviour.
type Config struct {
	// Workers is the number of concurrent send goroutines.
	// Default: 3.
	Workers int
	// MaxRetries is the maximum number of retries per chat on transient errors.
	// Default: 2.
	MaxRetries int
	// InitialDelay is the duration before the first retry.
	// Default: 2s.
	InitialDelay time.Duration
	// MaxDelay is the upper bound on retry backoff.
	// Default: 30s.
	MaxDelay time.Duration
}

// Defaults returns a Config with sensible defaults applied for any zero fields.
func (c Config) Defaults() Config {
	if c.Workers <= 0 {
		c.Workers = 3
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 2
	}
	if c.InitialDelay <= 0 {
		c.InitialDelay = 2 * time.Second
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = 30 * time.Second
	}
	return c
}

// PeerTarget is a chat to broadcast to, specified by its numeric chat ID.
// The caller is responsible for resolving human-readable identifiers
// (usernames, phone numbers) to chat IDs before constructing targets,
// using adapter.Context.ResolvePeerToID.
type PeerTarget struct {
	ChatID int64
}

// MakeSendFunc creates a SendFunc for a given chatID. The returned function
// is called by the broadcaster with the attempt number. Implementations
// should perform the actual message send and return any error.
type MakeSendFunc func(chatID int64) SendFunc

// Broadcaster sends messages to a list of PeerTargets using a bounded worker pool.
type Broadcaster struct {
	targets []PeerTarget
	cfg     Config
}

// New creates a Broadcaster for the given targets with the given configuration.
func New(targets []PeerTarget, cfg Config) *Broadcaster {
	return &Broadcaster{
		targets: targets,
		cfg:     cfg.Defaults(),
	}
}

// Targets returns a copy of the broadcast target list.
func (b *Broadcaster) Targets() []PeerTarget {
	cp := make([]PeerTarget, len(b.targets))
	copy(cp, b.targets)
	return cp
}

// Config returns the resolved configuration (with defaults applied).
func (b *Broadcaster) Config() Config {
	return b.cfg
}

// Run executes the broadcast. For each target, it calls makeSend to get a
// chat-specific SendFunc, then calls it with retry logic.
//
// Returns a BroadcastResult with per-chat outcomes. The result is fully
// populated when Run returns — callers do not need to synchronize access.
//
// Run blocks until all targets have been processed or the context is cancelled.
func (b *Broadcaster) Run(ctx context.Context, makeSend MakeSendFunc) *BroadcastResult {
	result := NewBroadcastResult(len(b.targets))
	if len(b.targets) == 0 {
		return result
	}

	// Feed targets into a channel for workers.
	ch := make(chan PeerTarget, len(b.targets))
	for _, target := range b.targets {
		ch <- target
	}
	close(ch)

	var wg sync.WaitGroup
	for i := 0; i < b.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range ch {
				// Check context before processing.
				select {
				case <-ctx.Done():
					result.RecordFailed(target.ChatID, ctx.Err())
					continue
				default:
				}

				chatID := target.ChatID
				send := makeSend(chatID)
				sent, attempts, err := sendWithRetry(ctx, send, b.cfg)

				if sent {
					result.RecordSent(chatID)
				} else if err != nil {
					result.RecordFailedWithAttempts(chatID, attempts, err)
				} else {
					// Non-retryable: skipped.
					result.RecordSkipped(chatID)
				}
			}
		}()
	}

	wg.Wait()
	return result
}

// TextSendConfig configures a text broadcast via adapter.Context.SendMessage.
type TextSendConfig struct {
	// Text is the message to broadcast.
	Text string
	// ParseMode overrides the client default. Empty means use client default.
	// Supported: "HTML", "MarkdownV2", "" (none).
	ParseMode string
	// Silent sends the message without notification.
	Silent bool
	// Background sends the message in the background.
	Background bool
}

// MakeTextSendFunc creates a MakeSendFunc that sends text via adapter.Context.SendMessage.
// Each chat gets its own call to ctx.SendMessage with a fresh tg.MessagesSendMessageRequest.
// ParseMode handling is delegated to the adapter layer.
func MakeTextSendFunc(ctx *adapter.Context, cfg TextSendConfig) MakeSendFunc {
	return func(chatID int64) SendFunc {
		return func(attempt int) error {
			req := &tg.MessagesSendMessageRequest{
				Message:    cfg.Text,
				Silent:     cfg.Silent,
				Background: cfg.Background,
			}
			_, err := ctx.SendMessage(chatID, req)
			return err
		}
	}
}
