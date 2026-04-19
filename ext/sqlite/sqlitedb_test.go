package sqlitedb

import (
	"fmt"
	"testing"
	"time"

	"github.com/pageton/gotg/storage"
)

func setupAdapter(t *testing.T) *SQLiteAdapter {
	t.Helper()
	dsn := fmt.Sprintf("file:gotg-sqlite-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	a, err := NewFromDSN(dsn)
	if err != nil {
		t.Fatalf("NewFromDSN: %v", err)
	}
	a.DB().SetMaxOpenConns(1)
	if err := a.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
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
	if got == nil || string(got.Data) != string(want.Data) {
		t.Fatalf("session mismatch: %+v", got)
	}

	want.Data = []byte("updated-key")
	if err := a.UpdateSession(want); err != nil {
		t.Fatalf("UpdateSession upsert: %v", err)
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
	p := &storage.Peer{
		ID:         42,
		AccessHash: 123,
		Type:       1,
		Username:   "alice",
		Usernames: storage.Usernames{
			{Username: "alice", Active: true, Editable: true},
			{Username: "alice_old", Active: false, Editable: false},
		},
		PhoneNumber: "12345",
		Language:    "en",
	}
	if err := a.SavePeer(p); err != nil {
		t.Fatalf("SavePeer: %v", err)
	}

	byID, err := a.GetPeerByID(42)
	if err != nil || byID == nil || byID.Username != "alice" {
		t.Fatalf("GetPeerByID mismatch: peer=%+v err=%v", byID, err)
	}
	if len(byID.Usernames) != 2 {
		t.Fatalf("expected 2 usernames, got %+v", byID.Usernames)
	}

	byUsername, err := a.GetPeerByUsername("alice")
	if err != nil || byUsername == nil || byUsername.ID != 42 {
		t.Fatalf("GetPeerByUsername primary mismatch: peer=%+v err=%v", byUsername, err)
	}

	byAlias, err := a.GetPeerByUsername("alice_old")
	if err != nil || byAlias == nil || byAlias.ID != 42 {
		t.Fatalf("GetPeerByUsername alias mismatch: peer=%+v err=%v", byAlias, err)
	}

	byPhone, err := a.GetPeerByPhoneNumber("12345")
	if err != nil || byPhone == nil || byPhone.ID != 42 {
		t.Fatalf("GetPeerByPhoneNumber mismatch: peer=%+v err=%v", byPhone, err)
	}
}

func TestConvState(t *testing.T) {
	a := setupAdapter(t)

	state := &storage.ConvState{
		Key:       "100:200",
		ChatID:    100,
		UserID:    200,
		Step:      "ask_name",
		Payload:   []byte(`{"name":"test"}`),
		ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Microsecond),
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := a.SaveConvState(state); err != nil {
		t.Fatalf("SaveConvState: %v", err)
	}

	got, err := a.LoadConvState("100:200")
	if err != nil || got == nil || got.Step != "ask_name" {
		t.Fatalf("LoadConvState mismatch: state=%+v err=%v", got, err)
	}

	states, err := a.ListConvStates()
	if err != nil || len(states) != 1 {
		t.Fatalf("ListConvStates mismatch: len=%d err=%v", len(states), err)
	}

	if err := a.DeleteConvState("100:200"); err != nil {
		t.Fatalf("DeleteConvState: %v", err)
	}
	got, err = a.LoadConvState("100:200")
	if err != nil || got != nil {
		t.Fatalf("expected nil after delete: state=%+v err=%v", got, err)
	}
}

func TestSessionNameIsolation(t *testing.T) {
	dsn := fmt.Sprintf("file:gotg-sqlite-scope-%d?mode=memory&cache=shared", time.Now().UnixNano())
	a1, err := NewFromDSN(dsn, SessionName("bot1"))
	if err != nil {
		t.Fatalf("NewFromDSN bot1: %v", err)
	}
	defer a1.Close()
	a1.DB().SetMaxOpenConns(1)

	a2, err := NewFromDSN(dsn, SessionName("bot2"))
	if err != nil {
		t.Fatalf("NewFromDSN bot2: %v", err)
	}
	defer a2.Close()
	a2.DB().SetMaxOpenConns(1)

	if err := a1.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate bot1: %v", err)
	}

	if err := a1.UpdateSession(&storage.Session{Version: 1, Data: []byte("one")}); err != nil {
		t.Fatalf("UpdateSession bot1: %v", err)
	}
	if err := a2.UpdateSession(&storage.Session{Version: 1, Data: []byte("two")}); err != nil {
		t.Fatalf("UpdateSession bot2: %v", err)
	}

	s1, err := a1.GetSession(1)
	if err != nil || s1 == nil || string(s1.Data) != "one" {
		t.Fatalf("bot1 session mismatch: session=%+v err=%v", s1, err)
	}
	s2, err := a2.GetSession(1)
	if err != nil || s2 == nil || string(s2.Data) != "two" {
		t.Fatalf("bot2 session mismatch: session=%+v err=%v", s2, err)
	}

	if err := a1.SavePeer(&storage.Peer{ID: 99, Username: "shared_name"}); err != nil {
		t.Fatalf("SavePeer bot1: %v", err)
	}
	p1, err := a1.GetPeerByID(99)
	if err != nil || p1 == nil {
		t.Fatalf("GetPeerByID bot1: peer=%+v err=%v", p1, err)
	}
	p2, err := a2.GetPeerByID(99)
	if err != nil || p2 != nil {
		t.Fatalf("GetPeerByID bot2 expected nil: peer=%+v err=%v", p2, err)
	}
}
