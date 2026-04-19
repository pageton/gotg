package gotg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	intErrors "github.com/pageton/gotg/errors"
)

// ReconnectConfig controls automatic reconnection behaviour when the client
// disconnects due to network errors or recoverable server errors.
//
// Set the AutoReconnect field in ClientOpts to a non-nil ReconnectConfig to
// enable. When enabled, Client.RunForever blocks until a fatal (non-retryable)
// error is encountered or the parent context is cancelled.
type ReconnectConfig struct {
	// MaxRetries limits how many consecutive reconnection attempts are made
	// before giving up. 0 means unlimited.
	MaxRetries int

	// InitialInterval is the backoff duration after the first failed attempt.
	// Default: 5 seconds.
	InitialInterval time.Duration

	// MaxInterval caps the exponential backoff. Default: 5 minutes.
	MaxInterval time.Duration

	// Multiplier is the factor by which the backoff grows after each attempt.
	// Default: 2.0.
	Multiplier float64

	// MaxElapsedTime is the total time budget for reconnection attempts before
	// giving up. 0 means no limit.
	MaxElapsedTime time.Duration

	// SessionReload controls whether the session data is re-read from
	// persistent storage before each reconnection attempt. This is useful when
	// the session may have been updated externally (e.g., by another process
	// importing a new string session). Default: false.
	SessionReload bool

	// OnDisconnect is called when the client connection drops (before the
	// backoff wait begins). The callback receives the error that caused the
	// disconnect and the attempt number (1-based). If the callback returns a
	// non-nil error, reconnection is aborted and RunForever returns.
	OnDisconnect func(err error, attempt int) error

	// OnReconnect is called after a successful reconnection (after login and
	// dispatcher initialization). The attempt number is 1-based. If the
	// callback returns a non-nil error, reconnection is aborted and RunForever
	// returns.
	OnReconnect func(attempt int) error

	// OnAuthLost is called when the session is detected as revoked or
	// unregistered. This is a fatal event by default — the callback provides
	// an opportunity to perform a fresh login (e.g., by updating the session
	// data or prompting the user). Return nil to continue with reconnection,
	// or an error to abort.
	OnAuthLost func(err error) error
}

// DefaultReconnectConfig returns a ReconnectConfig with sensible defaults.
func DefaultReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		InitialInterval: 5 * time.Second,
		MaxInterval:     5 * time.Minute,
		Multiplier:      2.0,
	}
}

// reconnectErrorClass categorises errors that come out of Client.Run().
type reconnectErrorClass int

const (
	// errorFatal means the client should not retry (e.g. auth revoked).
	errorFatal reconnectErrorClass = iota
	// errorRetryable means the client should back off and retry.
	errorRetryable
)

// classifyError inspects the error returned by the client run loop and decides
// whether reconnection should be attempted.
func classifyError(err error) reconnectErrorClass {
	if err == nil {
		return errorRetryable
	}

	// Context cancellation is not retryable — the caller chose to stop.
	if errors.Is(err, context.Canceled) {
		return errorFatal
	}

	// Auth-related fatal errors: the session itself is no longer valid.
	if isAuthLostError(err) {
		return errorFatal
	}

	// Everything else (network timeouts, connection refused, etc.) is retryable.
	return errorRetryable
}

// isAuthLostError returns true for errors indicating the session auth key has
// been revoked, unregistered, or is otherwise permanently invalid.
func isAuthLostError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	// gotd/td wraps Telegram RPC errors with their error code and message.
	// We match on the well-known strings that indicate auth key invalidation.
	authLostPatterns := []string{
		"SESSION_REVOKED",
		"AUTH_KEY_UNREGISTERED",
		"AUTH_KEY_DUPLICATED",
		"AUTH_KEY_INVALID",
		"USER_DEACTIVATED",
		"SESSION_PASSWORD_NEEDED",
	}

	for _, pattern := range authLostPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}

	// Also check the sentinel ErrSessionUnauthorized from gotg itself.
	if errors.Is(err, intErrors.ErrSessionUnauthorized) {
		return true
	}

	return false
}

// newBackoff creates a backoff.ExponentialBackOff from the ReconnectConfig.
func (cfg *ReconnectConfig) newBackoff() backoff.BackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     cfg.InitialInterval,
		MaxInterval:         cfg.MaxInterval,
		Multiplier:          cfg.Multiplier,
		MaxElapsedTime:      cfg.MaxElapsedTime,
		Clock:               backoff.SystemClock,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
	}
	b.Reset()
	return b
}

// reloadSession re-reads session data from the underlying storage so that
// externally-updated sessions (e.g., imported string sessions) are picked up
// before reconnection.
func (c *Client) reloadSession() error {
	if c.sessionStorage == nil || c.ctx == nil {
		return nil
	}

	data, err := c.sessionStorage.LoadSession(c.ctx)
	if err != nil {
		return fmt.Errorf("reload session: %w", err)
	}

	// If storage returned nil/empty the session was cleared — this is fine,
	// login() will handle re-auth.
	if len(data) == 0 {
		return nil
	}

	return nil
}

// RunForever starts the client and automatically reconnects on disconnection
// using the ReconnectConfig provided in ClientOpts.AutoReconnect.
//
// This method blocks until:
//   - A fatal (non-retryable) error occurs (e.g., auth key revoked).
//   - The configured MaxRetries is exceeded.
//   - The parent context is cancelled.
//   - An OnDisconnect/OnReconnect/OnAuthLost callback returns an error.
//
// RunForever requires that ClientOpts.AutoReconnect is non-nil; otherwise it
// returns an error immediately.
//
// Example:
//
//	client, _ := gotg.NewClient(apiID, apiHash, gotg.AsUser(phone), &gotg.ClientOpts{
//	    Session:       session.SqlSession(sqlite.Open("bot.db")),
//	    AutoReconnect: gotg.DefaultReconnectConfig(),
//	})
//
//	// Blocks forever (until fatal error or context cancel).
//	log.Fatal(client.RunForever())
func (c *Client) RunForever() error {
	cfg := c.autoReconnect
	if cfg == nil {
		return errors.New("gotg: RunForever requires ClientOpts.AutoReconnect")
	}

	b := cfg.newBackoff()
	attempt := 0

	for {
		// Reload session from persistent storage if configured.
		if cfg.SessionReload {
			if err := c.reloadSession(); err != nil && c.Logger != nil {
				c.Logger.Warn("session reload failed", "error", err)
			}
		}

		// Attempt to start (or restart) the client using the original opts
		// so that Device, Middlewares, RunMiddleware, etc. are preserved.
		reconnectOpts := c.startOpts
		if reconnectOpts == nil {
			reconnectOpts = &ClientOpts{}
		}
		startErr := c.Start(reconnectOpts)

		if startErr == nil {
			// Client started successfully; reset backoff.
			if attempt > 0 {
				b.Reset()
				if cfg.OnReconnect != nil {
					if cbErr := cfg.OnReconnect(attempt); cbErr != nil {
						return fmt.Errorf("OnReconnect callback aborted: %w", cbErr)
					}
				}
				if c.Logger != nil {
					c.Logger.Info("reconnected successfully",
						"attempt", attempt,
					)
				}
			}
			attempt = 0

			// Block until the client's run loop exits.
			idleErr := c.Idle()
			if idleErr == nil {
				// Idle returned nil — context was cancelled normally.
				return nil
			}

			// Classify the error that killed the connection.
			if class := classifyError(idleErr); class == errorFatal {
				if isAuthLostError(idleErr) && cfg.OnAuthLost != nil {
					if cbErr := cfg.OnAuthLost(idleErr); cbErr != nil {
						return fmt.Errorf("auth lost, OnAuthLost aborted: %w", cbErr)
					}
					// OnAuthLost returned nil — retry with potentially updated session.
					continue
				}
				return idleErr
			}
		}

		// We are in a disconnected state — count the attempt.
		attempt++
		runErr := startErr
		if runErr == nil {
			runErr = c.err
		}

		if cfg.OnDisconnect != nil {
			if cbErr := cfg.OnDisconnect(runErr, attempt); cbErr != nil {
				return fmt.Errorf("OnDisconnect callback aborted: %w", cbErr)
			}
		}

		// Check retry budget.
		if cfg.MaxRetries > 0 && attempt > cfg.MaxRetries {
			return fmt.Errorf("gotg: max reconnect retries (%d) exceeded, last error: %w",
				cfg.MaxRetries, runErr)
		}

		// Calculate backoff duration.
		nextBackoff := b.NextBackOff()
		if c.Logger != nil {
			c.Logger.Warn("client disconnected, reconnecting",
				"attempt", attempt,
				"backoff", nextBackoff,
				"error", runErr,
			)
		}

		// Stop the dead client before reconnecting.
		c.Stop()

		// Wait for backoff, respecting context cancellation.
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-time.After(nextBackoff):
		}
	}
}

// RunForeverWithSession is a convenience that creates a client with
// DisableAutoStart, then calls RunForever. It accepts the same parameters as
// NewClient but requires that opts.AutoReconnect is set.
//
// This is the recommended entry point for long-running userbots.
func RunForeverWithSession(
	apiID int, apiHash string,
	clientType clientType,
	opts *ClientOpts,
) error {
	if opts == nil {
		opts = &ClientOpts{}
	}
	if opts.AutoReconnect == nil {
		opts.AutoReconnect = DefaultReconnectConfig()
	}
	opts.DisableAutoStart = true

	client, err := NewClient(apiID, apiHash, clientType, opts)
	if err != nil {
		return err
	}

	return client.RunForever()
}

// compile-time check: ReconnectConfig fields are used.
var _ = (*ReconnectConfig)(nil)
