package adapter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
)

// ResolveInputPeerByID resolves a peer ID to an InputPeer.
func (ctx *Context) ResolveInputPeerByID(id int64) (tg.InputPeerClass, error) {
	return functions.ResolveInputPeerByID(ctx.Context, ctx.Raw, ctx.PeerStorage, id)
}

// ResolvePeerByID resolves a peer ID to a storage.Peer with metadata.
func (ctx *Context) ResolvePeerByID(id int64) *storage.Peer {
	_, _ = ctx.ResolveInputPeerByID(id)
	peer := ctx.PeerStorage.GetPeerByID(id)
	if peer != nil && peer.ID != 0 {
		return peer
	}
	ID := constant.TDLibPeerID(id)
	if ID.IsUser() {
		ID.Channel(id)
		peer = ctx.ResolvePeerByID(int64(ID))
		if peer != nil && peer.ID != 0 {
			return peer
		}
		ID.Chat(id)
		peer = ctx.ResolvePeerByID(int64(ID))
		if peer != nil && peer.ID != 0 {
			return peer
		}

	}
	return peer
}

// ResolvePeerToID resolves a human-readable peer identifier to a TDLib-format
// chat ID that can be passed directly to SendMessage, SendMedia, etc.
//
// The identifier is resolved in the following order:
//
//   - Numeric string (e.g. "12345678", "-1001234567890"): parsed as int64
//     and looked up via ResolveInputPeerByID (3-tier: cache → DB → RPC).
//
//   - Username (e.g. "@alice", "alice"): looked up via PeerStorage username
//     index (O(1) in-memory → DB → contacts.ResolveUsername RPC fallback).
//
//   - Phone number (e.g. "+1234567890", "1234567890"): looked up via
//     PeerStorage phone index (O(1) in-memory → DB → contacts.ResolvePhone
//     RPC fallback).
//
// Returns the TDLib-format peer ID (suitable for SendMessage chatID parameter)
// or an error if the peer cannot be resolved.
//
// Example:
//
//	// All of these resolve automatically:
//	id, _ := ctx.ResolvePeerToID("@alice")       // username
//	id, _ := ctx.ResolvePeerToID("+1234567890")   // phone
//	id, _ := ctx.ResolvePeerToID("-1001234567890") // channel ID
//	ctx.SendMessage(id, &tg.MessagesSendMessageRequest{Message: "Hello!"})
func (ctx *Context) ResolvePeerToID(identifier string) (int64, error) {
	if ctx.PeerStorage == nil {
		return 0, fmt.Errorf("peer storage not available")
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return 0, fmt.Errorf("empty identifier")
	}

	// Phone numbers start with + — skip numeric parsing and go to phone lookup.
	if strings.HasPrefix(identifier, "+") {
		peer := ctx.PeerStorage.GetPeerByPhoneNumber(identifier)
		if peer != nil && peer.ID != 0 {
			return peer.ID, nil
		}
		return 0, fmt.Errorf("peer not found by phone: %s", identifier)
	}

	// 1. Try numeric ID (may include negative TDLib prefix like -100...).
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		// Verify the peer exists via the 3-tier resolver.
		if _, resolveErr := ctx.ResolveInputPeerByID(id); resolveErr == nil {
			return id, nil
		}
		// Even if RPC resolution fails, the ID is valid — return it.
		// The send function will attempt resolution again.
		return id, nil
	}

	// 2. Try username lookup (strip leading @).
	if strings.HasPrefix(identifier, "@") || isLikelyUsername(identifier) {
		username := strings.TrimPrefix(identifier, "@")
		peer := ctx.PeerStorage.GetPeerByUsername(username)
		if peer != nil && peer.ID != 0 {
			return peer.ID, nil
		}
		return 0, fmt.Errorf("peer not found by username: @%s", username)
	}

	// 3. No match — unknown identifier format.
	return 0, fmt.Errorf("peer not found: %s", identifier)
}

// isLikelyUsername returns true for strings that look like Telegram usernames
// (alphanumeric + underscores, 5-32 chars, not purely numeric, not starting with +).
func isLikelyUsername(s string) bool {
	if len(s) < 5 || len(s) > 32 {
		return false
	}
	if s[0] == '+' {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	// Purely numeric strings are IDs, not usernames.
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return false
	}
	return true
}
