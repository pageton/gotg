package conv

import (
	"fmt"
	"sync"
	"time"

	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
)

// Filter determines whether an incoming message should be delivered to the waiter.
type Filter func(*types.Message) bool

// DefaultFilter accepts any non-nil message.
func DefaultFilter(msg *types.Message) bool { return msg != nil }

// Key identifies a conversation via chat+user pair.
type Key struct {
	ChatID int64
	UserID int64
}

func (k Key) String() string {
	return fmt.Sprintf("%d:%d", k.ChatID, k.UserID)
}

// StepHandler handles a conversation step response.
// The handler receives the State, and should return an error or nil.
type StepHandler func(state *State) error

// Manager keeps track of pending asks per peer.
type Manager struct {
	mu             sync.RWMutex
	steps          map[string]StepHandler
	filters        map[string]Filter
	storage        *storage.PeerStorage
	defaultTimeout time.Duration
}

// NewManager creates a new conversation manager.
func NewManager(p *storage.PeerStorage, defaultTimeout time.Duration) *Manager {
	if defaultTimeout <= 0 {
		defaultTimeout = 30 * time.Second
	}
	return &Manager{
		storage:        p,
		steps:          make(map[string]StepHandler),
		filters:        make(map[string]Filter),
		defaultTimeout: defaultTimeout,
	}
}

// RegisterStep registers a handler for a specific conversation step.
func (m *Manager) RegisterStep(name string, handler StepHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps[name] = handler
}

// GetStepHandler retrieves a registered step handler by name.
func (m *Manager) GetStepHandler(name string) (StepHandler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.steps[name]
	return h, ok
}

// SetStateWithOpts persists a conversation state for the given key.
func (m *Manager) SetStateWithOpts(key Key, step string, payload []byte, timeout time.Duration, filter Filter) error {
	m.mu.Lock()
	if filter != nil {
		m.filters[key.String()] = filter
	} else {
		delete(m.filters, key.String())
	}
	m.mu.Unlock()

	if m.storage == nil {
		return nil
	}
	state := &storage.ConvState{
		Key:     storage.ConvKey(key.ChatID, key.UserID),
		ChatID:  key.ChatID,
		UserID:  key.UserID,
		Step:    step,
		Payload: payload,
	}
	if timeout > 0 {
		state.ExpiresAt = time.Now().Add(timeout)
	}
	return m.storage.SaveConversationState(state)
}

// SetState persists a conversation state for the given key.
func (m *Manager) SetState(key Key, step string, payload []byte, timeout ...time.Duration) error {
	var t time.Duration
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return m.SetStateWithOpts(key, step, payload, t, nil)
}

// GetFilter retrieves the current filter for a conversation key.
func (m *Manager) GetFilter(key Key) Filter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.filters[key.String()]
}

// ClearFilter removes the filter for a conversation key.
func (m *Manager) ClearFilter(key Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.filters, key.String())
}

// LoadState retrieves and wraps the conversation state for a key.
func (m *Manager) LoadState(key Key, msg *types.Message, update any) (*State, error) {
	raw, err := m.GetState(key)
	if err != nil || raw == nil {
		return nil, err
	}
	state := newState(raw, m)
	state.Message = msg
	state.Update = update
	return state, nil
}

// SaveState persists a State object.
func (m *Manager) SaveState(state *State) error {
	return state.save()
}

// ClearState removes the conversation state for the given key.
func (m *Manager) ClearState(key Key) error {
	if m.storage == nil {
		return nil
	}
	return m.storage.DeleteConversationState(storage.ConvKey(key.ChatID, key.UserID))
}

// GetState retrieves the conversation state for the given key.
// Returns nil if no state exists or if it has expired.
func (m *Manager) GetState(key Key) (*storage.ConvState, error) {
	if m.storage == nil {
		return nil, nil
	}
	state, err := m.storage.LoadConversationState(storage.ConvKey(key.ChatID, key.UserID))
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil
	}
	// Check expiry
	if !state.ExpiresAt.IsZero() && time.Now().After(state.ExpiresAt) {
		_ = m.storage.DeleteConversationState(state.Key)
		return nil, nil
	}
	return state, nil
}
