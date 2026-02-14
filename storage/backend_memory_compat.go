package storage

import "sync"

type memoryAdapterCompat struct {
	mu         sync.RWMutex
	sessions   map[int]*Session
	peers      map[int64]*Peer
	convStates map[string]*ConvState
}

func (m *memoryAdapterCompat) GetSession(version int) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[version]
	if !ok {
		return nil, nil
	}
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	return &cp, nil
}

func (m *memoryAdapterCompat) UpdateSession(s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	m.sessions[s.Version] = &cp
	return nil
}

func (m *memoryAdapterCompat) SavePeer(p *Peer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.peers[p.ID] = &cp
	return nil
}

func (m *memoryAdapterCompat) GetPeerByID(id int64) (*Peer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.peers[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (m *memoryAdapterCompat) GetPeerByUsername(username string) (*Peer, error) {
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

func (m *memoryAdapterCompat) SaveConvState(state *ConvState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *state
	cp.Payload = append([]byte(nil), state.Payload...)
	m.convStates[state.Key] = &cp
	return nil
}

func (m *memoryAdapterCompat) LoadConvState(key string) (*ConvState, error) {
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

func (m *memoryAdapterCompat) DeleteConvState(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.convStates, key)
	return nil
}

func (m *memoryAdapterCompat) ListConvStates() ([]ConvState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	states := make([]ConvState, 0, len(m.convStates))
	for _, s := range m.convStates {
		cp := *s
		cp.Payload = append([]byte(nil), s.Payload...)
		states = append(states, cp)
	}
	return states, nil
}

func (m *memoryAdapterCompat) AutoMigrate() error { return nil }

func (m *memoryAdapterCompat) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[int]*Session)
	m.peers = make(map[int64]*Peer)
	m.convStates = make(map[string]*ConvState)
	return nil
}
