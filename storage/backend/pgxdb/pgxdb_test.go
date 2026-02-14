package pgxdb

import (
	"testing"
	"time"

	"github.com/pageton/gotg/storage"
)

const testDSN = "postgresql:///dbot?host=/tmp"

func setupAdapter(t *testing.T) *PgxAdapter {
	t.Helper()
	a, err := NewFromDSN(testDSN)
	if err != nil {
		t.Fatalf("NewFromDSN: %v", err)
	}
	if err := a.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	t.Cleanup(func() {
		a.pool.Exec(a.ctx, "DROP TABLE IF EXISTS sessions, peers, conv_states")
		a.Close()
	})
	return a
}

func TestSession(t *testing.T) {
	a := setupAdapter(t)

	s, err := a.GetSession(1)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if s != nil {
		t.Fatal("expected nil session")
	}

	want := &storage.Session{Version: 1, Data: []byte("test-auth-key")}
	if err := a.UpdateSession(want); err != nil {
		t.Fatalf("UpdateSession: %v", err)
	}

	got, err := a.GetSession(1)
	if err != nil {
		t.Fatalf("GetSession after update: %v", err)
	}
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if string(got.Data) != string(want.Data) {
		t.Fatalf("data mismatch: got %q, want %q", got.Data, want.Data)
	}

	want.Data = []byte("updated-key")
	if err := a.UpdateSession(want); err != nil {
		t.Fatalf("UpdateSession (upsert): %v", err)
	}
	got, err = a.GetSession(1)
	if err != nil {
		t.Fatalf("GetSession after upsert: %v", err)
	}
	if string(got.Data) != "updated-key" {
		t.Fatalf("upsert failed: got %q", got.Data)
	}
}

func TestPeers(t *testing.T) {
	a := setupAdapter(t)

	got, err := a.GetPeerByID(42)
	if err != nil {
		t.Fatalf("GetPeerByID: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil peer")
	}

	p := &storage.Peer{ID: 42, AccessHash: 123, Type: 1, Username: "alice", Language: "en"}
	if err := a.SavePeer(p); err != nil {
		t.Fatalf("SavePeer: %v", err)
	}

	got, err = a.GetPeerByID(42)
	if err != nil {
		t.Fatalf("GetPeerByID after save: %v", err)
	}
	if got == nil || got.Username != "alice" {
		t.Fatalf("peer mismatch: %+v", got)
	}

	got, err = a.GetPeerByUsername("alice")
	if err != nil {
		t.Fatalf("GetPeerByUsername: %v", err)
	}
	if got == nil || got.ID != 42 {
		t.Fatalf("username lookup mismatch: %+v", got)
	}

	got, err = a.GetPeerByUsername("nonexistent")
	if err != nil {
		t.Fatalf("GetPeerByUsername nonexistent: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for nonexistent username, got %+v", got)
	}

	p.Language = "ar"
	if err := a.SavePeer(p); err != nil {
		t.Fatalf("SavePeer upsert: %v", err)
	}
	got, _ = a.GetPeerByID(42)
	if got.Language != "ar" {
		t.Fatalf("upsert failed: language=%q", got.Language)
	}
}

func TestConvState(t *testing.T) {
	a := setupAdapter(t)

	got, err := a.LoadConvState("100:200")
	if err != nil {
		t.Fatalf("LoadConvState: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil conv state")
	}

	state := &storage.ConvState{
		Key:       "100:200",
		ChatID:    100,
		UserID:    200,
		Step:      "ask_name",
		Payload:   []byte(`{"name":"test"}`),
		ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Microsecond),
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := a.SaveConvState(state); err != nil {
		t.Fatalf("SaveConvState: %v", err)
	}

	got, err = a.LoadConvState("100:200")
	if err != nil {
		t.Fatalf("LoadConvState after save: %v", err)
	}
	if got == nil || got.Step != "ask_name" {
		t.Fatalf("conv state mismatch: %+v", got)
	}
	if string(got.Payload) != `{"name":"test"}` {
		t.Fatalf("payload mismatch: %q", got.Payload)
	}

	states, err := a.ListConvStates()
	if err != nil {
		t.Fatalf("ListConvStates: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 state, got %d", len(states))
	}

	if err := a.DeleteConvState("100:200"); err != nil {
		t.Fatalf("DeleteConvState: %v", err)
	}
	got, _ = a.LoadConvState("100:200")
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}
