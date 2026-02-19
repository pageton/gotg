package main

import (
	"os"
	"testing"
	"time"

	"github.com/pageton/gotg/storage"
)

func setupJson(t *testing.T) *JsonAdapter {
	t.Helper()
	path := t.TempDir() + "/test.json"
	a, err := NewJsonAdapter(path)
	if err != nil {
		t.Fatalf("NewJsonAdapter: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	return a
}

func TestJsonSession(t *testing.T) {
	a := setupJson(t)

	s, err := a.GetSession(1)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if s != nil {
		t.Fatal("expected nil session")
	}

	want := &storage.Session{Version: 1, Data: []byte("auth-key-data")}
	if err := a.UpdateSession(want); err != nil {
		t.Fatalf("UpdateSession: %v", err)
	}

	got, err := a.GetSession(1)
	if err != nil {
		t.Fatalf("GetSession after update: %v", err)
	}
	if got == nil || string(got.Data) != "auth-key-data" {
		t.Fatalf("data mismatch: %+v", got)
	}

	want.Data = []byte("updated-key")
	if err := a.UpdateSession(want); err != nil {
		t.Fatalf("UpdateSession overwrite: %v", err)
	}
	got, _ = a.GetSession(1)
	if string(got.Data) != "updated-key" {
		t.Fatalf("overwrite failed: got %q", got.Data)
	}
}

func TestJsonPeers(t *testing.T) {
	a := setupJson(t)

	got, err := a.GetPeerByID(42)
	if err != nil {
		t.Fatalf("GetPeerByID: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil peer")
	}

	p := &storage.Peer{ID: 42, AccessHash: 999, Type: 1, Username: "bob", Language: "en"}
	if err := a.SavePeer(p); err != nil {
		t.Fatalf("SavePeer: %v", err)
	}

	got, _ = a.GetPeerByID(42)
	if got == nil || got.Username != "bob" {
		t.Fatalf("peer mismatch: %+v", got)
	}

	got, _ = a.GetPeerByUsername("bob")
	if got == nil || got.ID != 42 {
		t.Fatalf("username lookup: %+v", got)
	}

	got, _ = a.GetPeerByUsername("nobody")
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}

	p.Language = "ar"
	a.SavePeer(p)
	got, _ = a.GetPeerByID(42)
	if got.Language != "ar" {
		t.Fatalf("upsert failed: %q", got.Language)
	}
}

func TestJsonConvState(t *testing.T) {
	a := setupJson(t)

	got, err := a.LoadConvState("1:2")
	if err != nil {
		t.Fatalf("LoadConvState: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}

	state := &storage.ConvState{
		Key:       "1:2",
		ChatID:    1,
		UserID:    2,
		Step:      "ask_name",
		Payload:   []byte(`{"q":"hi"}`),
		ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
		CreatedAt: time.Now().Truncate(time.Second),
	}
	if err := a.SaveConvState(state); err != nil {
		t.Fatalf("SaveConvState: %v", err)
	}

	got, _ = a.LoadConvState("1:2")
	if got == nil || got.Step != "ask_name" || string(got.Payload) != `{"q":"hi"}` {
		t.Fatalf("conv mismatch: %+v", got)
	}

	states, _ := a.ListConvStates()
	if len(states) != 1 {
		t.Fatalf("expected 1, got %d", len(states))
	}

	a.DeleteConvState("1:2")
	got, _ = a.LoadConvState("1:2")
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestJsonPersistence(t *testing.T) {
	path := t.TempDir() + "/persist.json"

	a, _ := NewJsonAdapter(path)
	a.SavePeer(&storage.Peer{ID: 1, Username: "alice"})
	a.UpdateSession(&storage.Session{Version: 1, Data: []byte("key")})
	a.Close()

	a2, err := NewJsonAdapter(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	p, _ := a2.GetPeerByID(1)
	if p == nil || p.Username != "alice" {
		t.Fatalf("peer not persisted: %+v", p)
	}

	s, _ := a2.GetSession(1)
	if s == nil || string(s.Data) != "key" {
		t.Fatalf("session not persisted: %+v", s)
	}
}
