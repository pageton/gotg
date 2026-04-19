package gotg

import (
	"testing"

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
	// We can't verify full connection (no real server), but we verify no nil-deref.
	defer client.Stop()

	// The call will fail to connect (no real Telegram server), but must not panic.
	_ = client.Start(nil)

	// Verify that passing nil didn't nil-deref — the client attempted a start.
	if client.Client == nil {
		t.Fatal("telegram client should be initialized after Start(nil)")
	}
}
