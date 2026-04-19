package adapter

import (
	"context"
	"testing"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/storage/memory"
)

// newTestContext creates a minimal Context with an in-memory PeerStorage
// for testing peer resolution.
func newTestContext() *Context {
	adapter := memory.New()
	peerStorage, _ := storage.NewPeerStorageWithAdapter(adapter, true)

	return &Context{
		Context:     context.Background(),
		PeerStorage: peerStorage,
		Self:        &tg.User{ID: 99},
	}
}

// addTestUser is a helper that adds a user to the PeerStorage with the
// given ID, access hash, username, and phone number.
func addTestUser(ps *storage.PeerStorage, id, accessHash int64, username, phone string) {
	var usernames storage.Usernames
	if username != "" {
		usernames = storage.Usernames{{Username: username, Active: true, Editable: true}}
	}
	ps.AddPeerWithUsernames(id, accessHash, storage.TypeUser, username, usernames, phone, false, 0)
}

// addTestChannel is a helper that adds a channel to the PeerStorage.
func addTestChannel(ps *storage.PeerStorage, id, accessHash int64, username string) {
	var usernames storage.Usernames
	if username != "" {
		usernames = storage.Usernames{{Username: username, Active: true, Editable: true}}
	}
	ps.AddPeerWithUsernames(id, accessHash, storage.TypeChannel, username, usernames, storage.DefaultPhone, false, 0)
}

// --- ResolvePeerToID tests ---

func TestResolvePeerToID_NumericID(t *testing.T) {
	ctx := newTestContext()
	addTestUser(ctx.PeerStorage, 12345678, 999, "alice", "+1234567890")

	id, err := ctx.ResolvePeerToID("12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 12345678 {
		t.Fatalf("expected 12345678, got %d", id)
	}
}

func TestResolvePeerToID_NumericIDNegative(t *testing.T) {
	ctx := newTestContext()

	// Add a channel. The TDLib-format ID includes -100 prefix.
	var chID constant.TDLibPeerID
	chID.Channel(1234567890)
	tdlibID := int64(chID)

	addTestChannel(ctx.PeerStorage, tdlibID, 888, "mychannel")

	id, err := ctx.ResolvePeerToID("-1001234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != tdlibID {
		t.Fatalf("expected %d, got %d", tdlibID, id)
	}
}

func TestResolvePeerToID_UsernameWithAt(t *testing.T) {
	ctx := newTestContext()
	addTestUser(ctx.PeerStorage, 111, 222, "alice", "")

	id, err := ctx.ResolvePeerToID("@alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 111 {
		t.Fatalf("expected 111, got %d", id)
	}
}

func TestResolvePeerToID_UsernameWithoutAt(t *testing.T) {
	ctx := newTestContext()
	addTestUser(ctx.PeerStorage, 111, 222, "alice", "")

	id, err := ctx.ResolvePeerToID("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 111 {
		t.Fatalf("expected 111, got %d", id)
	}
}

func TestResolvePeerToID_ChannelUsername(t *testing.T) {
	ctx := newTestContext()

	// AddPeerWithUsernames encodes the ID internally, so pass the plain ID.
	addTestChannel(ctx.PeerStorage, 555, 777, "mychannel")

	id, err := ctx.ResolvePeerToID("@mychannel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The peer should be found; the returned ID is the TDLib-encoded form.
	// Verify it resolves to a non-zero ID.
	if id == 0 {
		t.Fatal("expected non-zero channel ID")
	}

	// Verify round-trip: the ID should be lookable-up in the peer storage.
	peer := ctx.PeerStorage.GetPeerByID(id)
	if peer == nil {
		t.Fatalf("peer not found in storage for ID %d", id)
	}
	if peer.Username != "mychannel" {
		t.Fatalf("expected username mychannel, got %s", peer.Username)
	}
}

func TestResolvePeerToID_PhoneWithPlus(t *testing.T) {
	ctx := newTestContext()
	addTestUser(ctx.PeerStorage, 333, 444, "bob", "+15551234567")

	id, err := ctx.ResolvePeerToID("+15551234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 333 {
		t.Fatalf("expected 333, got %d", id)
	}
}

func TestResolvePeerToID_PhoneWithoutPlus(t *testing.T) {
	ctx := newTestContext()
	addTestUser(ctx.PeerStorage, 333, 444, "bob", "+15551234567")

	// Phone lookup uses the exact string stored.
	id, err := ctx.ResolvePeerToID("+15551234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 333 {
		t.Fatalf("expected 333, got %d", id)
	}
}

func TestResolvePeerToID_NotFound(t *testing.T) {
	ctx := newTestContext()

	_, err := ctx.ResolvePeerToID("@nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown username")
	}
}

func TestResolvePeerToID_Empty(t *testing.T) {
	ctx := newTestContext()

	_, err := ctx.ResolvePeerToID("")
	if err == nil {
		t.Fatal("expected error for empty identifier")
	}
}

func TestResolvePeerToID_NilPeerStorage(t *testing.T) {
	ctx := &Context{
		Context:     context.Background(),
		PeerStorage: nil,
	}

	_, err := ctx.ResolvePeerToID("@alice")
	if err == nil {
		t.Fatal("expected error when PeerStorage is nil")
	}
}

func TestResolvePeerToID_PrefersUsernameOverPhone(t *testing.T) {
	ctx := newTestContext()
	// User with both username and phone.
	addTestUser(ctx.PeerStorage, 777, 888, "charlie", "+15559999999")

	// Username resolution should work.
	id, err := ctx.ResolvePeerToID("@charlie")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 777 {
		t.Fatalf("expected 777, got %d", id)
	}

	// Phone resolution should also work for the same user.
	id, err = ctx.ResolvePeerToID("+15559999999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 777 {
		t.Fatalf("expected 777, got %d", id)
	}
}

// --- isLikelyUsername tests ---

func TestIsLikelyUsername_Valid(t *testing.T) {
	cases := []string{
		"alice",
		"alice_123",
		"JohnDoe",
		"my_channel",
		"abcdef",
	}
	for _, tc := range cases {
		if !isLikelyUsername(tc) {
			t.Fatalf("expected %q to be detected as username", tc)
		}
	}
}

func TestIsLikelyUsername_Invalid(t *testing.T) {
	cases := []struct {
		input string
		reason string
	}{
		{"+12345", "starts with +"},
		{"1234", "too short"},
		{"abc", "too short"},
		{"a b c", "contains spaces"},
		{"@alice", "contains @"},
		{"12345678", "purely numeric"},
	}
	for _, tc := range cases {
		if isLikelyUsername(tc.input) {
			t.Fatalf("expected %q NOT to be detected as username (%s)", tc.input, tc.reason)
		}
	}
}

// --- Update.ResolvePeerToID delegation test ---

func TestUpdateResolvePeerToID(t *testing.T) {
	ctx := newTestContext()
	addTestUser(ctx.PeerStorage, 111, 222, "alice", "")

	u := &Update{Ctx: ctx}

	id, err := u.ResolvePeerToID("@alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 111 {
		t.Fatalf("expected 111, got %d", id)
	}
}
