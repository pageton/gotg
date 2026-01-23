package conversation

import (
	"context"
	"fmt"
	"sync"
	"time"

	mtp_errors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
)

var (
	ErrTimeout   = mtp_errors.ErrConversationTimeout
	ErrCancelled = mtp_errors.ErrConversationCancelled
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

// Options configures how the manager waits for the next reply.
type Options struct {
	Timeout time.Duration
	Filter  Filter
	Step    string
	Payload []byte
}

// Manager keeps track of pending asks per peer.
type Manager struct {
	mu             sync.RWMutex
	waiters        map[string]*waiter
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
		waiters:        make(map[string]*waiter),
		defaultTimeout: defaultTimeout,
	}
}

// WaitResponse registers a waiter for the given key and blocks until a reply arrives or timeout occurs.
func (m *Manager) WaitResponse(ctx context.Context, key Key, opts Options) (*types.Message, error) {
	if key.ChatID == 0 || key.UserID == 0 {
		return nil, fmt.Errorf("conversation: invalid key %s", key)
	}
	if opts.Filter == nil {
		opts.Filter = DefaultFilter
	}
	if opts.Timeout <= 0 {
		opts.Timeout = m.defaultTimeout
	}

	w := &waiter{
		key:     key,
		filter:  opts.Filter,
		updates: make(chan *types.Message, 1),
		done:    make(chan struct{}),
		timeout: opts.Timeout,
	}
	m.registerWaiter(w)
	defer m.unregisterWaiter(w.key)

	m.persistState(key, opts)
	defer m.forgetState(key)

	timer := time.NewTimer(opts.Timeout)
	defer timer.Stop()

	select {
	case msg := <-w.updates:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, mtp_errors.ErrConversationTimeout
	case <-w.done:
		return nil, mtp_errors.ErrConversationCancelled
	}
}

// Route attempts to deliver a message to a registered waiter.
func (m *Manager) Route(key Key, msg *types.Message) bool {
	if msg == nil {
		return false
	}
	m.mu.RLock()
	w := m.waiters[key.String()]
	m.mu.RUnlock()
	if w == nil {
		return false
	}
	if !w.filter(msg) {
		return false
	}
	select {
	case w.updates <- msg:
	default:
		return false
	}
	return true
}

// ActiveKeys returns currently registered conversation keys.
func (m *Manager) ActiveKeys() []Key {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]Key, 0, len(m.waiters))
	for _, w := range m.waiters {
		keys = append(keys, w.key)
	}
	return keys
}

func (m *Manager) registerWaiter(w *waiter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.waiters[w.key.String()] = w
}

func (m *Manager) unregisterWaiter(key Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w, ok := m.waiters[key.String()]; ok {
		close(w.done)
		delete(m.waiters, key.String())
	}
}

func (m *Manager) persistState(key Key, opts Options) {
	if m.storage == nil {
		return
	}
	state := &storage.ConversationState{
		Key:     storage.ConversationKey(key.ChatID, key.UserID),
		ChatID:  key.ChatID,
		UserID:  key.UserID,
		Step:    opts.Step,
		Payload: opts.Payload,
	}
	if opts.Timeout > 0 {
		state.ExpiresAt = time.Now().Add(opts.Timeout)
	}
	_ = m.storage.SaveConversationState(state)
}

func (m *Manager) forgetState(key Key) {
	if m.storage == nil {
		return
	}
	_ = m.storage.DeleteConversationState(storage.ConversationKey(key.ChatID, key.UserID))
}

type waiter struct {
	key     Key
	filter  Filter
	updates chan *types.Message
	done    chan struct{}
	timeout time.Duration
}
