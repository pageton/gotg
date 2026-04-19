package session

import (
	"context"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/storage/memory"
)

// NewSessionStorage creates a session storage for Telegram client authentication.
//
// It creates both a PeerStorage (for peer caching) and a SessionStorage
// (for persisting auth session data like auth keys, DC ID, etc.).
//
// Parameters:
//   - ctx: Context for session operations
//   - sessionType: Session type to load (WithAdapter, SimpleSession, StringSession)
//   - inMemory: If true, only in-memory storage is used (no database persistence)
//
// Returns:
//   - PeerStorage: For caching peers (users, chats, channels)
//   - SessionStorage: For storing Telegram session data
//   - error: If session loading fails
func NewSessionStorage(ctx context.Context, sessionType SessionConstructor, inMemory bool) (*storage.PeerStorage, telegram.SessionStorage, error) {
	name, data, err := sessionType.loadSession()
	if err != nil {
		return nil, nil, err
	}

	switch n := name.(type) {
	case sessionNameAdapter:
		peerStorage, err := storage.NewPeerStorageWithAdapter(n.adapter, false)
		if err != nil {
			return nil, nil, err
		}
		return peerStorage, &SessionStorage{
			data:        peerStorage.GetSession().Data,
			peerStorage: peerStorage,
		}, nil

	case sessionNameAdapterWithData:
		// String session + adapter: adapter handles peers and conv state.
		// If the adapter already has a session (from a previous run),
		// prefer it over the string — the string may be stale.
		peerStorage, err := storage.NewPeerStorageWithAdapter(n.adapter, false)
		if err != nil {
			return nil, nil, err
		}
		data := n.data
		if existing := peerStorage.GetSession(); existing != nil && len(existing.Data) > 0 {
			data = existing.Data
		}
		return peerStorage, &SessionStorage{
			data:        data,
			peerStorage: peerStorage,
		}, nil

	default:
		// All string-based sessions without adapter use in-memory storage.
		adapter := memory.New()
		peerStorage, err := storage.NewPeerStorageWithAdapter(adapter, inMemory)
		if err != nil {
			return nil, nil, err
		}
		if inMemory {
			s := session.StorageMemory{}
			if err := s.StoreSession(ctx, data); err != nil {
				return nil, nil, err
			}
			return peerStorage, &s, nil
		}
		return peerStorage, &SessionStorage{
			data:        data,
			peerStorage: peerStorage,
		}, nil
	}
}
