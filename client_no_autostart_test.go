package gotg

import (
	"testing"
	"time"

	"github.com/pageton/gotg/session"
)

func TestNewClientDisableAutoStart(t *testing.T) {
	client, err := NewClient(1, "test-hash", AsBot("test-token"), &ClientOpts{
		InMemory:         true,
		Session:          session.SimpleSession(),
		DisableAutoStart: true,
		DisableCopyright: true,
	})
	if err != nil {
		t.Fatalf("NewClient returned error with DisableAutoStart=true: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient returned nil client")
	}

	if client.running {
		t.Fatal("client should not be running when DisableAutoStart=true")
	}

	if client.Client != nil {
		t.Fatal("telegram client should not be initialized before Start")
	}

	client.Stop()
}

func TestStartWithNilOpts(t *testing.T) {
	origOpts := &ClientOpts{
		InMemory:         true,
		Session:          session.SimpleSession(),
		DisableAutoStart: true,
		DisableCopyright: true,
	}

	client, err := NewClient(1, "test-hash", AsBot("test-token"), origOpts)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if client.startOpts != origOpts {
		t.Fatal("client.startOpts should be set to the original opts from NewClient")
	}

	// Start with nil opts should fall back to startOpts without panicking.
	// We can't fully test connection (no real server), so we verify no panic occurs.
	defer client.Stop()

	// Use a separate goroutine to detect panics without blocking test.
	panicked := make(chan interface{}, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicked <- r
			}
			close(panicked)
		}()
		// Start will block trying to connect, but should not panic.
		// We don't wait for it to complete since there's no real server.
		_ = client.Start(nil)
	}()

	select {
	case r, ok := <-panicked:
		if ok && r != nil {
			t.Fatalf("Start(nil) panicked: %v", r)
		}
		// No panic - test passes
	case <-time.After(2 * time.Second):
		// Start is blocking (expected without real server), but didn't panic
		// This is acceptable - we're just testing for nil-deref/panic
	}
}
