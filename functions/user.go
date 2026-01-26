package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// GetUser returns tg.User of provided user ID.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - userID: The user ID to get info for
//
// Returns user information or an error.
func GetUser(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, userID int64) (*tg.User, error) {
	peerUser := GetInputPeerClassFromID(p, userID)
	if peerUser == nil {
		return nil, errors.ErrPeerNotFound
	}

	switch peer := peerUser.(type) {
	case *tg.InputPeerUser:
		users, err := raw.UsersGetUsers(ctx, []tg.InputUserClass{
			&tg.InputUser{
				UserID:     peer.UserID,
				AccessHash: peer.AccessHash,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, errors.ErrPeerNotFound
		}
		user, ok := users[0].(*tg.User)
		if !ok {
			return nil, errors.ErrPeerNotFound
		}
		return user, nil
	default:
		return nil, errors.ErrNotUser
	}
}

// GetFullUser returns tg.UserFull of provided user ID.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - userID: The user ID to get full info for
//
// Returns full user information or an error.
func GetFullUser(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, userID int64) (*tg.UserFull, error) {
	peerUser := GetInputPeerClassFromID(p, userID)
	if peerUser == nil {
		return nil, errors.ErrPeerNotFound
	}

	switch peer := peerUser.(type) {
	case *tg.InputPeerUser:
		user, err := raw.UsersGetFullUser(ctx, &tg.InputUser{
			UserID:     peer.UserID,
			AccessHash: peer.AccessHash,
		})
		if err != nil {
			return nil, err
		}
		return &user.FullUser, nil
	default:
		return nil, errors.ErrNotUser
	}
}
