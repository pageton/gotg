package broadcast

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsRetryable_NilError(t *testing.T) {
	assert.False(t, isRetryable(nil))
}

func TestIsRetryable_ContextCanceled(t *testing.T) {
	assert.False(t, isRetryable(context.Canceled))
}

func TestIsRetryable_ContextDeadlineExceeded(t *testing.T) {
	assert.False(t, isRetryable(context.DeadlineExceeded))
}

func TestIsRetryable_NonRetryableTelegramErrors(t *testing.T) {
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
	}
	for _, code := range nonRetryable {
		err := fmt.Errorf("rpc error: %s", code)
		assert.False(t, isRetryable(err), "expected %s to be non-retryable", code)
	}
}

func TestIsRetryable_FloodWait(t *testing.T) {
	err := fmt.Errorf("rpc error: FLOOD_WAIT_5")
	assert.True(t, isRetryable(err))
}

func TestIsRetryable_NetworkError(t *testing.T) {
	err := errors.New("connection reset by peer")
	assert.True(t, isRetryable(err))
}

func TestIsRetryable_Timeout(t *testing.T) {
	err := errors.New("i/o timeout")
	assert.True(t, isRetryable(err))
}

func TestIsRetryable_PeerNotFound(t *testing.T) {
	// Peer not found is non-retryable — the peer won't appear on retry.
	err := errors.New("peer not found")
	assert.False(t, isRetryable(err))
}

func TestSendWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	var calls []int
	sender := func(attempt int) error {
		calls = append(calls, attempt)
		return nil
	}
	cfg := Config{MaxRetries: 3, InitialDelay: 0, MaxDelay: 0}
	sent, attempts, err := sendWithRetry(context.Background(), sender, cfg)
	assert.True(t, sent)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
	assert.Len(t, calls, 1)
}

func TestSendWithRetry_SuccessOnSecondAttempt(t *testing.T) {
	var calls []int
	sender := func(attempt int) error {
		calls = append(calls, attempt)
		if attempt == 0 {
			return errors.New("temporary network error")
		}
		return nil
	}
	cfg := Config{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}
	sent, attempts, err := sendWithRetry(context.Background(), sender, cfg)
	assert.True(t, sent)
	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.Len(t, calls, 2)
}

func TestSendWithRetry_NonRetryableError(t *testing.T) {
	var calls []int
	sender := func(attempt int) error {
		calls = append(calls, attempt)
		return errors.New("CHAT_WRITE_FORBIDDEN")
	}
	cfg := Config{MaxRetries: 3, InitialDelay: 0, MaxDelay: 0}
	sent, attempts, err := sendWithRetry(context.Background(), sender, cfg)
	assert.False(t, sent)
	assert.NoError(t, err) // non-retryable is a skip, not a failure error
	assert.Equal(t, 1, attempts)
	assert.Len(t, calls, 1)
}

func TestSendWithRetry_ExhaustsRetries(t *testing.T) {
	var calls []int
	sender := func(attempt int) error {
		calls = append(calls, attempt)
		return errors.New("connection reset")
	}
	cfg := Config{MaxRetries: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}
	sent, attempts, err := sendWithRetry(context.Background(), sender, cfg)
	assert.False(t, sent)
	assert.Error(t, err)
	assert.Equal(t, 3, attempts) // initial + 2 retries = 3 attempts
	assert.Len(t, calls, 3)
}

func TestSendWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	sender := func(attempt int) error {
		return ctx.Err() // returns context.Canceled
	}
	cfg := Config{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}
	sent, attempts, err := sendWithRetry(ctx, sender, cfg)
	assert.False(t, sent)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, attempts)
}