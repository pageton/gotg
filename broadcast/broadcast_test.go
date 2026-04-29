package broadcast

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBroadcaster_AllSucceed(t *testing.T) {
	targets := []PeerTarget{
		{ChatID: 100},
		{ChatID: 200},
		{ChatID: 300},
	}

	b := New(targets, Config{Workers: 2, MaxRetries: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond})

	var sentIDs []int64
	var mu sync.Mutex
	result := b.Run(context.Background(), func(chatID int64) SendFunc {
		return func(attempt int) error {
			mu.Lock()
			sentIDs = append(sentIDs, chatID)
			mu.Unlock()
			return nil
		}
	})

	assert.Equal(t, 3, result.Sent())
	assert.Equal(t, 0, result.Failed())
	assert.Equal(t, 0, result.Skipped())
	assert.Equal(t, 3, result.Total())
	assert.ElementsMatch(t, []int64{100, 200, 300}, sentIDs)
}

func TestBroadcaster_MixedResults(t *testing.T) {
	targets := []PeerTarget{
		{ChatID: 100}, // succeeds
		{ChatID: 200}, // non-retryable error (skip)
		{ChatID: 300}, // transient error, then succeeds
		{ChatID: 400}, // always fails (transient)
	}

	cfg := Config{Workers: 2, MaxRetries: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}
	b := New(targets, cfg)

	result := b.Run(context.Background(), func(chatID int64) SendFunc {
		var attempts atomic.Int32
		return func(attempt int) error {
			switch chatID {
			case 100:
				return nil
			case 200:
				return errors.New("CHAT_WRITE_FORBIDDEN")
			case 300:
				if attempts.Add(1) == 1 {
					return errors.New("timeout")
				}
				return nil
			case 400:
				return errors.New("connection reset")
			default:
				return nil
			}
		}
	})

	assert.Equal(t, 2, result.Sent(), "100 and 300 should succeed")
	assert.Equal(t, 1, result.Skipped(), "200 should be skipped")
	assert.Equal(t, 1, result.Failed(), "400 should fail after retries")
	assert.Equal(t, 4, result.Total())
}

func TestBroadcaster_ContextCancellation(t *testing.T) {
	targets := []PeerTarget{
		{ChatID: 100},
		{ChatID: 200},
	}

	cfg := Config{Workers: 1, MaxRetries: 0, InitialDelay: 0, MaxDelay: 0}
	b := New(targets, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before any work

	result := b.Run(ctx, func(chatID int64) SendFunc {
		return func(attempt int) error {
			return ctx.Err()
		}
	})

	assert.Equal(t, 2, result.Failed())
}

func TestBroadcaster_EmptyTargets(t *testing.T) {
	b := New(nil, Config{})
	result := b.Run(context.Background(), func(chatID int64) SendFunc {
		return func(attempt int) error { return nil }
	})
	assert.Equal(t, 0, result.Total())
	assert.Equal(t, 0, result.Sent())
}

func TestBroadcaster_Defaults(t *testing.T) {
	b := New([]PeerTarget{{ChatID: 1}}, Config{})
	cfg := b.Config()
	assert.Equal(t, 3, cfg.Workers)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, 2*time.Second, cfg.InitialDelay)
	assert.Equal(t, 30*time.Second, cfg.MaxDelay)
}