package broadcast

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// nonRetryablePatterns lists Telegram error codes that indicate a permanent
// failure for the target chat. Retrying will not change the outcome.
var nonRetryablePatterns = []string{
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
	"peer not found",
}

// isRetryable returns true for errors that may succeed on a subsequent attempt.
// Non-retryable errors include: context cancellation, auth/permission failures,
// and peer-not-found (the peer won't magically appear).
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := err.Error()
	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(msg, pattern) {
			return false
		}
	}
	return true
}

// newExponentialBackoff creates a backoff.ExponentialBackOff from the given
// parameters. It is initialized (Reset called) before returning.
func newExponentialBackoff(initialDelay, maxDelay time.Duration) *backoff.ExponentialBackOff {
	bo := &backoff.ExponentialBackOff{
		InitialInterval:     initialDelay,
		MaxInterval:         maxDelay,
		Multiplier:          2.0,
		MaxElapsedTime:      0, // retry until MaxRetries exhausted
		Clock:               backoff.SystemClock,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
	}
	bo.Reset()
	return bo
}

// sendWithRetry attempts to send via the provided SendFunc with exponential
// backoff on retryable errors. Returns:
//   - (true, attempts, nil) on success
//   - (false, attempts, nil) if skipped (non-retryable error)
//   - (false, attempts, err) after exhausting retries
func sendWithRetry(ctx context.Context, send SendFunc, cfg Config) (sent bool, attempts int, lastErr error) {
	cfg = cfg.Defaults()
	bo := newExponentialBackoff(cfg.InitialDelay, cfg.MaxDelay)

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		attempts = attempt + 1
		if attempt > 0 {
			next := bo.NextBackOff()
			select {
			case <-ctx.Done():
				return false, attempts, ctx.Err()
			case <-time.After(next):
			}
		}

		err := send(attempt)
		if err == nil {
			return true, attempts, nil
		}

		lastErr = err

		// Context cancellation is never retryable regardless of what
		// isRetryable says — bail immediately.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, attempts, err
		}

		if !isRetryable(err) {
			// Non-retryable: skip this chat, no error propagated.
			return false, attempts, nil
		}
	}

	return false, attempts, lastErr
}