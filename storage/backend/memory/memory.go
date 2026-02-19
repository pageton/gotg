// Package memory provides an in-memory Backend implementation.
// Useful for testing, ephemeral bots, and development.
//
// Usage:
//
//	b := memory.New()
package memory

import (
	"sync"

	"github.com/pageton/gotg/storage"
)

type MemoryAdapter struct {
	mu         sync.RWMutex
	sessions   map[int]*storage.Session
	peers      map[int64]*storage.Peer
	convStates map[string]*storage.ConvState
}

// New creates a new MemoryBackend.
func New() *MemoryAdapter {
	return &MemoryAdapter{
		sessions:   make(map[int]*storage.Session),
		peers:      make(map[int64]*storage.Peer),
		convStates: make(map[string]*storage.ConvState),
	}
}

func (m *MemoryAdapter) GetSession(version int) (*storage.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[version]
	if !ok {
		return nil, nil
	}
	// Return a copy to avoid data races on the slice.
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	return &cp, nil
}

func (m *MemoryAdapter) UpdateSession(s *storage.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	m.sessions[s.Version] = &cp
	return nil
}

func (m *MemoryAdapter) SavePeer(p *storage.Peer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.peers[p.ID] = &cp
	return nil
}

func (m *MemoryAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.peers[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (m *MemoryAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.peers {
		if p.Username == username {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MemoryAdapter) SaveConvState(state *storage.ConvState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *state
	cp.Payload = append([]byte(nil), state.Payload...)
	m.convStates[state.Key] = &cp
	return nil
}

func (m *MemoryAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.convStates[key]
	if !ok {
		return nil, nil
	}
	cp := *s
	cp.Payload = append([]byte(nil), s.Payload...)
	return &cp, nil
}

func (m *MemoryAdapter) DeleteConvState(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.convStates, key)
	return nil
}

func (m *MemoryAdapter) ListConvStates() ([]storage.ConvState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	states := make([]storage.ConvState, 0, len(m.convStates))
	for _, s := range m.convStates {
		cp := *s
		cp.Payload = append([]byte(nil), s.Payload...)
		states = append(states, cp)
	}
	return states, nil
}

func (m *MemoryAdapter) AutoMigrate() error { return nil }

func (m *MemoryAdapter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[int]*storage.Session)
	m.peers = make(map[int64]*storage.Peer)
	m.convStates = make(map[string]*storage.ConvState)
	return nil
}
