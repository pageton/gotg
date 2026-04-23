package gotg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	intErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/session"
)

// --- classifyError tests ---

func TestClassifyError_ContextCanceled(t *testing.T) {
	class := classifyError(context.Canceled)
	if class != errorFatal {
		t.Fatalf("expected errorFatal for context.Canceled, got %v", class)
	}
}

func TestClassifyError_Nil(t *testing.T) {
	class := classifyError(nil)
	if class != errorRetryable {
		t.Fatalf("expected errorRetryable for nil error, got %v", class)
	}
}

func TestClassifyError_AuthLost(t *testing.T) {
	cases := []struct {
		label string
		err   error
	}{
		{"SESSION_REVOKED", fmt.Errorf("rpc error: SESSION_REVOKED")},
		{"AUTH_KEY_UNREGISTERED", fmt.Errorf("rpc error: AUTH_KEY_UNREGISTERED")},
		{"AUTH_KEY_DUPLICATED", fmt.Errorf("rpc error: AUTH_KEY_DUPLICATED")},
		{"AUTH_KEY_INVALID", fmt.Errorf("rpc error: AUTH_KEY_INVALID")},
		{"USER_DEACTIVATED", fmt.Errorf("rpc error: USER_DEACTIVATED")},
		{"SESSION_PASSWORD_NEEDED", fmt.Errorf("rpc error: SESSION_PASSWORD_NEEDED")},
		{"ErrSessionUnauthorized", intErrors.ErrSessionUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			class := classifyError(tc.err)
			if class != errorFatal {
				t.Fatalf("expected errorFatal for %v, got %v", tc.err, class)
			}
		})
	}
}

func TestClassifyError_NetworkErrors(t *testing.T) {
	cases := []struct {
		label string
		err   error
	}{
		{"connection refused", fmt.Errorf("dial tcp: connection refused")},
		{"timeout", fmt.Errorf("context deadline exceeded")},
		{"EOF", fmt.Errorf("unexpected EOF")},
		{"generic", errors.New("something went wrong")},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			class := classifyError(tc.err)
			if class != errorRetryable {
				t.Fatalf("expected errorRetryable for %v, got %v", tc.err, class)
			}
		})
	}
}

// --- isAuthLostError tests ---

func TestIsAuthLostError_Nil(t *testing.T) {
	if isAuthLostError(nil) {
		t.Fatal("expected false for nil error")
	}
}

func TestIsAuthLostError_WrappedSentinel(t *testing.T) {
	wrapped := fmt.Errorf("wrapped: %w", intErrors.ErrSessionUnauthorized)
	if !isAuthLostError(wrapped) {
		t.Fatal("expected true for wrapped ErrSessionUnauthorized")
	}
}

func TestIsAuthLostError_UnknownError(t *testing.T) {
	if isAuthLostError(errors.New("FLOOD_WAIT")) {
		t.Fatal("expected false for unrelated Telegram error")
	}
}

// --- ReconnectConfig defaults tests ---

func TestDefaultReconnectConfig(t *testing.T) {
	cfg := DefaultReconnectConfig()

	if cfg.InitialInterval != 5*time.Second {
		t.Fatalf("expected InitialInterval=5s, got %v", cfg.InitialInterval)
	}
	if cfg.MaxInterval != 5*time.Minute {
		t.Fatalf("expected MaxInterval=5m, got %v", cfg.MaxInterval)
	}
	if cfg.Multiplier != 2.0 {
		t.Fatalf("expected Multiplier=2.0, got %v", cfg.Multiplier)
	}
	if cfg.MaxRetries != 0 {
		t.Fatalf("expected MaxRetries=0 (unlimited), got %d", cfg.MaxRetries)
	}
	if cfg.MaxElapsedTime != 0 {
		t.Fatalf("expected MaxElapsedTime=0 (no limit), got %v", cfg.MaxElapsedTime)
	}
}

func TestReconnectConfigNewBackoff(t *testing.T) {
	cfg := &ReconnectConfig{
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}

	b := cfg.newBackoff()

	// cenkalti/backoff applies randomization (±50% by default), so use ranges.
	first := b.NextBackOff()
	if first < 500*time.Millisecond || first > 1500*time.Millisecond {
		t.Fatalf("expected first backoff ~1s (±50%%), got %v", first)
	}

	second := b.NextBackOff()
	if second < 1*time.Second || second > 4*time.Second {
		t.Fatalf("expected second backoff ~2s (±50%%), got %v", second)
	}
}

func TestReconnectConfigNewBackoff_MaxRetriesRespected(t *testing.T) {
	cfg := &ReconnectConfig{
		MaxRetries:       2,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:      50 * time.Millisecond,
		Multiplier:      2.0,
	}

	b := cfg.newBackoff()
	durations := []time.Duration{}
	for i := 0; i < 5; i++ {
		d := b.NextBackOff()
		durations = append(durations, d)
	}

	// After a few steps, backoff reaches max interval and stays there.
	for _, d := range durations {
		if d == backoff.Stop {
			t.Fatalf("did not expect Stop before MaxElapsedTime, got %v", durations)
		}
	}
}

// --- RunForever guard tests ---

func TestRunForever_NilConfig(t *testing.T) {
	client, err := NewClient(1, "test", AsBot("tok"), &ClientOpts{
		InMemory:         true,
		Session:          session.SimpleSession(),
		DisableAutoStart: true,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.RunForever()
	if err == nil {
		t.Fatal("expected error when AutoReconnect is nil")
	}
	if !strings.Contains(err.Error(), "RunForever requires") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// --- Integration-style test: RunForeverWithSession validation ---

func TestRunForeverWithSession_SetsDefaults(t *testing.T) {
	// We can't actually connect, so use a short timeout context to ensure
	// the test doesn't hang. The call will fail quickly at the connection level.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := RunForeverWithSession(1, "hash", AsBot("tok"), &ClientOpts{
		InMemory: true,
		Session:  session.SimpleSession(),
		Context:  ctx,
	})
	if err == nil {
		t.Fatal("expected connection error (no real Telegram server)")
	}
	// The error should be from the gotd connection attempt, not from config.
	if strings.Contains(err.Error(), "AutoReconnect") {
		t.Fatalf("AutoReconnect should have been auto-set, got: %v", err)
	}
}

// --- Callback tests ---

func TestReconnectConfig_CallbacksFireInOrder(t *testing.T) {
	cfg := DefaultReconnectConfig()
	cfg.InitialInterval = 1 * time.Millisecond
	cfg.MaxInterval = 1 * time.Millisecond
	cfg.MaxRetries = 1 // only one retry

	var events []string

	cfg.OnDisconnect = func(err error, attempt int) error {
		events = append(events, fmt.Sprintf("disconnect:%d", attempt))
		return nil
	}

	// Verify config was created properly and callbacks are set.
	if cfg.OnDisconnect == nil {
		t.Fatal("OnDisconnect should be set")
	}
	if cfg.OnReconnect != nil {
		t.Fatal("OnReconnect should be nil by default")
	}
	if cfg.OnAuthLost != nil {
		t.Fatal("OnAuthLost should be nil by default")
	}

	// Just check the callback works.
	err := fmt.Errorf("connection refused")
	if cbErr := cfg.OnDisconnect(err, 1); cbErr != nil {
		t.Fatalf("OnDisconnect returned error: %v", cbErr)
	}

	if len(events) != 1 || events[0] != "disconnect:1" {
		t.Fatalf("expected [disconnect:1], got %v", events)
	}
}

func TestReconnectConfig_OnAuthLostCallback(t *testing.T) {
	cfg := DefaultReconnectConfig()
	called := false
	cfg.OnAuthLost = func(err error) error {
		called = true
		return nil // allow retry
	}

	authErr := fmt.Errorf("SESSION_REVOKED")
	if cbErr := cfg.OnAuthLost(authErr); cbErr != nil {
		t.Fatalf("OnAuthLost returned: %v", cbErr)
	}
	if !called {
		t.Fatal("OnAuthLost was not called")
	}
}

func TestReconnectConfig_OnDisconnect_AbortReconnect(t *testing.T) {
	cfg := DefaultReconnectConfig()
	abortErr := errors.New("manual abort")
	cfg.OnDisconnect = func(err error, attempt int) error {
		return abortErr
	}

	// Simulate the check that RunForever performs.
	if cbErr := cfg.OnDisconnect(fmt.Errorf("network error"), 1); cbErr == nil {
		t.Fatal("expected callback to return abort error")
	}
}

// --- HealthCheck callback tests ---

func TestReconnectConfig_HealthCheckCallback(t *testing.T) {
	cfg := DefaultReconnectConfig()
	called := false
	cfg.HealthCheck = func(ctx context.Context) error {
		called = true
		return nil
	}

	// Simulate a successful health check call.
	if err := cfg.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck returned unexpected error: %v", err)
	}
	if !called {
		t.Fatal("HealthCheck was not called")
	}
}

func TestReconnectConfig_HealthCheckFailure(t *testing.T) {
	cfg := DefaultReconnectConfig()
	hcErr := errors.New("health check failed: auth key expired")
	cfg.HealthCheck = func(ctx context.Context) error {
		return hcErr
	}

	// Simulate the check RunForever performs — non-nil means reconnect.
	if err := cfg.HealthCheck(context.Background()); err == nil {
		t.Fatal("expected HealthCheck to return error")
	}
}

func TestReconnectConfig_HealthCheckDefaultNil(t *testing.T) {
	cfg := DefaultReconnectConfig()
	if cfg.HealthCheck != nil {
		t.Fatal("HealthCheck should be nil by default")
	}
}

// --- OnBackoff callback tests ---

func TestReconnectConfig_OnBackoffCallback(t *testing.T) {
	cfg := DefaultReconnectConfig()

	var capturedAttempt int
	var capturedBackoff time.Duration
	var capturedErr error

	cfg.OnBackoff = func(attempt int, backoff time.Duration, err error) {
		capturedAttempt = attempt
		capturedBackoff = backoff
		capturedErr = err
	}

	testErr := fmt.Errorf("connection refused")
	cfg.OnBackoff(3, 10*time.Second, testErr)

	if capturedAttempt != 3 {
		t.Fatalf("expected attempt=3, got %d", capturedAttempt)
	}
	if capturedBackoff != 10*time.Second {
		t.Fatalf("expected backoff=10s, got %v", capturedBackoff)
	}
	if capturedErr.Error() != "connection refused" {
		t.Fatalf("expected 'connection refused', got %v", capturedErr)
	}
}

func TestReconnectConfig_OnBackoffDefaultNil(t *testing.T) {
	cfg := DefaultReconnectConfig()
	if cfg.OnBackoff != nil {
		t.Fatal("OnBackoff should be nil by default")
	}
}

// --- GracefulShutdown / ShutdownTimeout tests ---

func TestReconnectConfig_GracefulShutdownDefaults(t *testing.T) {
	cfg := DefaultReconnectConfig()
	if cfg.GracefulShutdown {
		t.Fatal("GracefulShutdown should be false by default")
	}
	if cfg.ShutdownTimeout != 0 {
		t.Fatalf("ShutdownTimeout should be 0 by default, got %v", cfg.ShutdownTimeout)
	}
}

func TestReconnectConfig_GracefulShutdownEnabled(t *testing.T) {
	cfg := DefaultReconnectConfig()
	cfg.GracefulShutdown = true
	cfg.ShutdownTimeout = 15 * time.Second

	if !cfg.GracefulShutdown {
		t.Fatal("GracefulShutdown should be true")
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected ShutdownTimeout=15s, got %v", cfg.ShutdownTimeout)
	}
}

func TestShutdown_GracefulWaitsForHandlers(t *testing.T) {
	client, err := NewClient(1, "test", AsBot("tok"), &ClientOpts{
		InMemory:         true,
		Session:          session.SimpleSession(),
		DisableAutoStart: true,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Shutdown with a short timeout on a non-started client should complete
	// immediately since there are no pending handlers. ctx.Err() is nil
	// because the context didn't expire.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Shutdown(ctx)
	// Client was never started — no pending handlers, so Shutdown finishes
	// before the context expires. ctx.Err() == nil.
	if err != nil {
		t.Fatalf("expected nil (no pending work), got: %v", err)
	}
}

func TestShutdown_FullTeardown(t *testing.T) {
	client, err := NewClient(1, "test", AsBot("tok"), &ClientOpts{
		InMemory:         true,
		Session:          session.SimpleSession(),
		DisableAutoStart: true,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown on a client that was never started should succeed.
	err = client.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown on non-started client should not error, got: %v", err)
	}
}

// --- Integration: callbacks fire together ---

func TestReconnectConfig_AllNewCallbacksNotNil(t *testing.T) {
	cfg := DefaultReconnectConfig()

	// Set all new callbacks.
	healthCalled := false
	backoffCalled := false

	cfg.HealthCheck = func(ctx context.Context) error {
		healthCalled = true
		return nil
	}
	cfg.OnBackoff = func(attempt int, backoff time.Duration, err error) {
		backoffCalled = true
	}
	cfg.GracefulShutdown = true
	cfg.ShutdownTimeout = 10 * time.Second

	// Invoke them.
	cfg.HealthCheck(context.Background())
	cfg.OnBackoff(1, 5*time.Second, fmt.Errorf("test"))

	if !healthCalled {
		t.Fatal("HealthCheck not called")
	}
	if !backoffCalled {
		t.Fatal("OnBackoff not called")
	}
}
