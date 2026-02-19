package adapter

import (
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
