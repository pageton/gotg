package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// GetPeerStories fetches all active stories for a given peer by chat ID.
//
// This resolves the peer from storage and calls the stories.getPeerStories method.
//
// Example:
//
//	peerStories, err := ctx.GetPeerStories(chatID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, story := range peerStories.Stories.Stories {
//	    log.Printf("Story: %v", story)
//	}
//
// Parameters:
//   - chatID: The chat/user ID whose stories to fetch
//
// Returns StoriesPeerStories containing the stories, or an error.
func (ctx *Context) GetPeerStories(chatID int64) (*tg.StoriesPeerStories, error) {
	return functions.GetPeerStories(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID)
}

// GetPeerStoriesByInputPeer fetches all active stories for a pre-resolved input peer.
//
// Use this when you already have a resolved InputPeerClass (e.g. from ResolveUsername).
//
// Example:
//
//	resolved, err := ctx.ResolveUsername("username")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	peerStories, err := ctx.GetPeerStoriesByInputPeer(resolved.GetInputPeer())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - peer: The resolved input peer
//
// Returns StoriesPeerStories containing the stories, or an error.
func (ctx *Context) GetPeerStoriesByInputPeer(peer tg.InputPeerClass) (*tg.StoriesPeerStories, error) {
	return functions.GetPeerStoriesByInputPeer(ctx.Context, ctx.Raw, peer)
}

// GetStoriesByID fetches specific stories by their IDs from a given peer.
//
// Example:
//
//	stories, err := ctx.GetStoriesByID(chatID, []int{1, 2, 3})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, story := range stories.Stories {
//	    log.Printf("Story: %v", story)
//	}
//
// Parameters:
//   - chatID: The chat/user ID whose stories to fetch
//   - ids: The story IDs to fetch
//
// Returns StoriesStories containing the requested stories, or an error.
func (ctx *Context) GetStoriesByID(chatID int64, ids []int) (*tg.StoriesStories, error) {
	return functions.GetStoriesByID(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, ids)
}

// GetStoriesByIDWithInputPeer fetches specific stories by their IDs using a pre-resolved input peer.
//
// Use this when you already have a resolved InputPeerClass (e.g. from ResolveUsername).
//
// Example:
//
//	resolved, err := ctx.ResolveUsername("username")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	stories, err := ctx.GetStoriesByIDWithInputPeer(resolved.GetInputPeer(), []int{42})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - peer: The resolved input peer
//   - ids: The story IDs to fetch
//
// Returns StoriesStories containing the requested stories, or an error.
func (ctx *Context) GetStoriesByIDWithInputPeer(peer tg.InputPeerClass, ids []int) (*tg.StoriesStories, error) {
	return functions.GetStoriesByIDWithInputPeer(ctx.Context, ctx.Raw, peer, ids)
}
