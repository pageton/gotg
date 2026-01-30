package session

import (
	"context"
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/pageton/gotg/storage"
)

// NewSessionStorage creates a session storage for Telegram client authentication.
//
// It creates both a PeerStorage (for peer caching) and a SessionStorage
// (for persisting auth session data like auth keys, DC ID, etc.).
//
// Parameters:
//   - ctx: Context for session operations
//   - sessionType: Session type to load (SqlSession, MemorySession, StringSession)
//   - inMemory: If true, only in-memory storage is used (no database persistence)
//
// Returns:
//   - PeerStorage: For caching peers (users, chats, channels)
//   - SessionStorage: For storing Telegram session data
//   - error: If session loading fails
//
// Example:
//
//	// Create session with SQLite database
//	peerStorage, sessionStorage, err := session.NewSessionStorage(
//	    ctx,
//	    session.SqlSession(sqlite.Open("telegram.db")),
//	    false,
//	)
//
//	// Create in-memory only session
//	peerStorage, sessionStorage, err := session.NewSessionStorage(
//	    ctx,
//	    session.MemorySession(),
//	    true,
//	)
func NewSessionStorage(ctx context.Context, sessionType SessionConstructor, inMemory bool) (*storage.PeerStorage, telegram.SessionStorage, error) {
	if sessionType == nil {
		sessionType = SimpleSession()
	}
	name, data, err := sessionType.loadSession()
	if err != nil {
		return nil, nil, err
	}
	if sessDialect, ok := name.(*sessionNameDialector); ok {
		peerStorage := storage.NewPeerStorage(sessDialect.dialector, false)
		return peerStorage, &SessionStorage{
			data:        peerStorage.GetSession().Data,
			peerStorage: peerStorage,
		}, nil
	}
	if name.(sessionNameString) == "" {
		name = sessionNameString("gotg")
	}
	peerStorage := storage.NewPeerStorage(sqlite.Open(fmt.Sprintf("%s.session", name)), inMemory)
	if inMemory {
		s := session.StorageMemory{}
		err := s.StoreSession(ctx, data)
		if err != nil {
			return nil, nil, err
		}
		return peerStorage, &s, nil
	}
	return peerStorage, &SessionStorage{
		data:        data,
		peerStorage: peerStorage,
	}, nil
}
