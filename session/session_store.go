package session

import (
	"context"
	"sync"

	"github.com/gotd/td/session"
	gotgErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// SessionStorage implements SessionStorage for file system as file
// stored in Path.
type SessionStorage struct {
	data        []byte
	peerStorage *storage.PeerStorage
	mux         sync.Mutex
}

type jsonData struct {
	Version int
	Data    session.Data
}

// LoadSession loads session from file.
func (f *SessionStorage) LoadSession(_ context.Context) ([]byte, error) {
	if f == nil {
		return nil, gotgErrors.ErrNilSessionStorage
	}

	f.mux.Lock()
	defer f.mux.Unlock()

	return f.data, nil
}

// StoreSession stores session to sqlite storage.
func (f *SessionStorage) StoreSession(_ context.Context, data []byte) error {
	if f == nil {
		return gotgErrors.ErrNilSessionStorage
	}
	f.mux.Lock()
	defer f.mux.Unlock()

	f.peerStorage.UpdateSession(&storage.Session{
		Version: storage.LatestVersion,
		Data:    data,
	})
	return nil
}
