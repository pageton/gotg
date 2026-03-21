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
