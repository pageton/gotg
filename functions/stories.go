package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// GetPeerStories fetches all active stories for a given peer.
//
// This wraps the raw stories.getPeerStories MTProto method,
// resolving the peer from storage if needed.
//
// Example:
//
//	peerStories, err := functions.GetPeerStories(ctx, raw, peerStorage, chatID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, story := range peerStories.Stories.Stories {
//	    fmt.Printf("Story: %v\n", story)
//	}
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat/user ID whose stories to fetch
//
// Returns StoriesPeerStories containing the stories, or an error.
func GetPeerStories(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) (*tg.StoriesPeerStories, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
	}
	return raw.StoriesGetPeerStories(ctx, inputPeer)
}

// GetPeerStoriesByInputPeer fetches all active stories for a given input peer directly.
//
// Use this when you already have a resolved InputPeerClass (e.g. from ResolveUsername).
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peer: The resolved input peer
//
// Returns StoriesPeerStories containing the stories, or an error.
func GetPeerStoriesByInputPeer(ctx context.Context, raw *tg.Client, peer tg.InputPeerClass) (*tg.StoriesPeerStories, error) {
	return raw.StoriesGetPeerStories(ctx, peer)
}

// GetStoriesByID fetches specific stories by their IDs from a given peer.
//
// This wraps the raw stories.getStoriesByID MTProto method,
// resolving the peer from storage if needed.
//
// Example:
//
//	stories, err := functions.GetStoriesByID(ctx, raw, peerStorage, chatID, []int{1, 2, 3})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, story := range stories.Stories {
//	    fmt.Printf("Story ID: %v\n", story)
//	}
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat/user ID whose stories to fetch
//   - ids: The story IDs to fetch
//
// Returns StoriesStories containing the requested stories, or an error.
func GetStoriesByID(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, ids []int) (*tg.StoriesStories, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
	}
	return raw.StoriesGetStoriesByID(ctx, &tg.StoriesGetStoriesByIDRequest{
		Peer: inputPeer,
		ID:   ids,
	})
}

// GetStoriesByIDWithInputPeer fetches specific stories by their IDs using a pre-resolved input peer.
//
// Use this when you already have a resolved InputPeerClass (e.g. from ResolveUsername).
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peer: The resolved input peer
//   - ids: The story IDs to fetch
//
// Returns StoriesStories containing the requested stories, or an error.
func GetStoriesByIDWithInputPeer(ctx context.Context, raw *tg.Client, peer tg.InputPeerClass, ids []int) (*tg.StoriesStories, error) {
	return raw.StoriesGetStoriesByID(ctx, &tg.StoriesGetStoriesByIDRequest{
		Peer: peer,
		ID:   ids,
	})
}
