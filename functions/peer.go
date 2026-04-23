package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// GetChatIDFromPeer returns a chat/user id from the provided tg.PeerClass.
//
// Parameters:
//   - peer: The peer object to extract ID from
//
// Returns chat/user ID.
func GetChatIDFromPeer(peer tg.PeerClass) int64 {
	var ID constant.TDLibPeerID
	switch peer := peer.(type) {
	case *tg.PeerChannel:
		ID.Channel(peer.ChannelID)
		return int64(ID)
	case *tg.PeerUser:
		return peer.UserID
	case *tg.PeerChat:
		ID.Chat(peer.ChatID)
		return int64(ID)
	default:
		return 0
	}
}

// GetInputPeerClassFromID finds provided user id in session storage and returns it if found.
//
// Parameters:
//   - p: Peer storage to search in
//   - iD: The user/chat ID to look up
//
// Returns input peer class or nil if not found.
func GetInputPeerClassFromID(p *storage.PeerStorage, id int64) tg.InputPeerClass {
	if p == nil {
		return nil
	}
	peer := p.GetPeerByID(id)
	if peer == nil || peer.ID == 0 {
		return nil
	}
	switch storage.EntityType(peer.Type) {
	case storage.TypeUser:
		return &tg.InputPeerUser{
			UserID:     peer.ID,
			AccessHash: peer.AccessHash,
		}
	case storage.TypeChat:
		return &tg.InputPeerChat{
			ChatID: peer.GetID(),
		}
	case storage.TypeChannel:
		return &tg.InputPeerChannel{
			ChannelID:  peer.GetID(),
			AccessHash: peer.AccessHash,
		}
	}
	return nil
}

// SavePeersFromClassArray saves chat and user peers from a class array to storage.
//
// Parameters:
//   - p: Peer storage to save to
//   - cs: List of chat classes to save
//   - us: List of user classes to save
//
// Returns nothing.
func SavePeersFromClassArray(p *storage.PeerStorage, cs []tg.ChatClass, us []tg.UserClass) {
	if p == nil {
		return
	}
	for _, u := range us {
		u, ok := u.(*tg.User)
		if !ok {
			continue
		}
		p.AddPeerWithUsernames(u.ID, u.AccessHash, storage.TypeUser, strings.ToLower(u.Username), storage.ConvertUsernames(u.Usernames), u.Phone, u.Bot, storage.ExtractPhotoID(u.Photo))
	}
	for _, c := range cs {
		switch c := c.(type) {
		case *tg.Channel:
			p.AddPeerWithUsernames(c.ID, c.AccessHash, storage.TypeChannel, strings.ToLower(c.Username), storage.ConvertUsernames(c.Usernames), storage.DefaultPhone, false, 0)
		case *tg.Chat:
			p.AddPeer(c.ID, storage.DefaultAccessHash, storage.TypeChat, storage.DefaultUsername)
		}
	}
}

// ResolveInputPeerByID tries to resolve given id to InputPeer.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peerStorage: Peer storage for resolving peer references
//   - id: The peer ID to resolve
//
// Returns input peer class or error if peer could not be resolved.
func ResolveInputPeerByID(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, id int64) (tg.InputPeerClass, error) {
	if peerStorage == nil {
		return nil, errors.ErrPeerNotFound
	}
	peer := peerStorage.GetInputPeerByID(id)
	if _, isEmpty := peer.(*tg.InputPeerEmpty); !isEmpty {
		return peer, nil
	}

	ID := constant.TDLibPeerID(id)
	if ID.IsChannel() { //nolint:gocritic // ifElseChain: method-call conditions, not switchable
		return nil, errors.ErrPeerNotFound
	} else if ID.IsChat() {
		plainID := ID.ToPlain()
		return &tg.InputPeerChat{
			ChatID: plainID,
		}, nil
	} else if ID.IsUser() {
		plainID := ID.ToPlain()
		users, err := raw.UsersGetUsers(ctx, []tg.InputUserClass{
			&tg.InputUser{
				UserID:     plainID,
				AccessHash: 0,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user: %w", err)
		}
		user, ok := tg.UserClassArray(users).FirstAsNotEmpty()
		if ok {
			peerStorage.AddPeerWithUsernames(user.ID, user.AccessHash, storage.TypeUser, strings.ToLower(user.Username), storage.ConvertUsernames(user.Usernames), user.Phone, user.Bot, storage.ExtractPhotoID(user.Photo))
			return user.AsInputPeer(), nil
		}
		// Try to get from storage again, but this time with bot-api compatible ids
		if ID.IsUser() {
			ID.Channel(id)
			peer := peerStorage.GetInputPeerByID(int64(ID))
			if _, isEmpty := peer.(*tg.InputPeerEmpty); !isEmpty {
				return peer, nil
			}
			ID.Chat(id)
			peer = peerStorage.GetInputPeerByID(int64(ID))
			if _, isEmpty := peer.(*tg.InputPeerEmpty); !isEmpty {
				return peer, nil
			}
		}
	}

	return nil, errors.ErrPeerNotFound
}
