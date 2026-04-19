package main

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pageton/gotg/storage"
)

// This file demonstrates how to implement a custom storage.Adapter.
//
// To create your own adapter (MongoDB, BoltDB, BadgerDB, etc.),
// implement all methods of the storage.Adapter interface:
//
//   GetSession / UpdateSession      — Telegram auth session
//   SavePeer / GetPeerByID / GetPeerByUsername — peer cache
//   SaveConvState / LoadConvState / DeleteConvState / ListConvStates — conversations
//   AutoMigrate / Close             — lifecycle
//
// Then pass it to:
//   storage.NewPeerStorageWithAdapter(yourAdapter, false)
// or:
//   session.WithAdapter(yourAdapter)

// jsonData is the on-disk format.
type jsonData struct {
	Sessions   map[int]*storage.Session      `json:"sessions"`
	Peers      map[int64]*storage.Peer       `json:"peers"`
	ConvStates map[string]*storage.ConvState `json:"conv_states"`
}

// JsonAdapter persists all storage to a single JSON file.
// Good for prototyping, bad for production (no concurrent process safety).
type JsonAdapter struct {
	mu   sync.RWMutex
	path string
	db   jsonData
}

func NewJsonAdapter(path string) (*JsonAdapter, error) {
	j := &JsonAdapter{
		path: path,
		db: jsonData{
			Sessions:   make(map[int]*storage.Session),
			Peers:      make(map[int64]*storage.Peer),
			ConvStates: make(map[string]*storage.ConvState),
		},
	}

	buf, err := os.ReadFile(path)
	if err == nil && len(buf) > 0 {
		if err := json.Unmarshal(buf, &j.db); err != nil {
			return nil, err
		}
		if j.db.Sessions == nil {
			j.db.Sessions = make(map[int]*storage.Session)
		}
		if j.db.Peers == nil {
			j.db.Peers = make(map[int64]*storage.Peer)
		}
		if j.db.ConvStates == nil {
			j.db.ConvStates = make(map[string]*storage.ConvState)
		}
	}
	return j, nil
}

func (j *JsonAdapter) flush() error {
	buf, err := json.MarshalIndent(j.db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(j.path, buf, 0644)
}

// --- Session ---

func (j *JsonAdapter) GetSession(version int) (*storage.Session, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	s, ok := j.db.Sessions[version]
	if !ok {
		return nil, nil
	}
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	return &cp, nil
}

func (j *JsonAdapter) UpdateSession(s *storage.Session) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	j.db.Sessions[s.Version] = &cp
	return j.flush()
}

// --- Peers ---

func (j *JsonAdapter) SavePeer(p *storage.Peer) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	cp := *p
	j.db.Peers[p.ID] = &cp
	return j.flush()
}

func (j *JsonAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	p, ok := j.db.Peers[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (j *JsonAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	for _, p := range j.db.Peers {
		if p.Username == username {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}

func (j *JsonAdapter) GetPeerByPhoneNumber(phone string) (*storage.Peer, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	for _, p := range j.db.Peers {
		if p.PhoneNumber == phone {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}

// --- Conversation State ---

func (j *JsonAdapter) SaveConvState(state *storage.ConvState) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	cp := *state
	cp.Payload = append([]byte(nil), state.Payload...)
	j.db.ConvStates[state.Key] = &cp
	return j.flush()
}

func (j *JsonAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	s, ok := j.db.ConvStates[key]
	if !ok {
		return nil, nil
	}
	cp := *s
	cp.Payload = append([]byte(nil), s.Payload...)
	return &cp, nil
}

func (j *JsonAdapter) DeleteConvState(key string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.db.ConvStates, key)
	return j.flush()
}

func (j *JsonAdapter) ListConvStates() ([]storage.ConvState, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	states := make([]storage.ConvState, 0, len(j.db.ConvStates))
	for _, s := range j.db.ConvStates {
		cp := *s
		cp.Payload = append([]byte(nil), s.Payload...)
		states = append(states, cp)
	}
	return states, nil
}

// --- Lifecycle ---

func (j *JsonAdapter) AutoMigrate() error { return nil }

func (j *JsonAdapter) DeleteStalePeers(olderThan int64) (int64, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	var count int64
	for id, p := range j.db.Peers {
		if p.LastUpdated > 0 && p.LastUpdated < olderThan {
			delete(j.db.Peers, id)
			count++
		}
	}
	if count > 0 {
		return count, j.flush()
	}
	return 0, nil
}

func (j *JsonAdapter) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.flush()
}
