package conv

import (
	"log"
	"sync"
	"time"

	"github.com/bytedance/sonic"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
)

type NextOpts struct {
	Filter  Filter
	Timeout time.Duration
	Reply   bool
}

type EndOpts struct {
	Reply bool
}

type State struct {
	mu      sync.RWMutex
	raw     *storage.ConvState
	data    map[string]any
	dirty   bool
	manager *Manager
	Update  any
	Message *types.Message
	SendFn  func(text string) error
	ReplyFn func(text string) error
	MediaFn func(media tg.InputMediaClass, caption string) error
}

func newState(raw *storage.ConvState, m *Manager) *State {
	s := &State{
		raw:     raw,
		data:    make(map[string]any),
		manager: m,
	}
	if len(raw.Payload) > 0 {
		if err := sonic.Unmarshal(raw.Payload, &s.data); err != nil {
			log.Printf("conv: failed to unmarshal state payload for step %q: %v", raw.Step, err)
		}
	}
	return s
}

func (s *State) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *State) GetString(key string) string {
	v := s.Get(key)
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

func (s *State) GetInt(key string) int {
	v := s.Get(key)
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	}
	return 0
}

func (s *State) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.dirty = true
}

func (s *State) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	s.dirty = true
}

func (s *State) Step() string {
	return s.raw.Step
}

func (s *State) ChatID() int64 {
	return s.raw.ChatID
}

func (s *State) UserID() int64 {
	return s.raw.UserID
}

func (s *State) Key() Key {
	return Key{ChatID: s.raw.ChatID, UserID: s.raw.UserID}
}

func (s *State) Text() string {
	if s.Message != nil {
		return s.Message.Text
	}
	return ""
}

func (s *State) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.dirty {
		return nil
	}
	payload, err := sonic.Marshal(s.data)
	if err != nil {
		return err
	}
	s.raw.Payload = payload
	return s.manager.storage.SaveConversationState(s.raw)
}

func (s *State) Next(step string, text string, opts ...*NextOpts) error {
	s.raw.Step = step
	if err := s.save(); err != nil {
		return err
	}

	var timeout time.Duration
	var filter Filter
	var reply bool
	if len(opts) > 0 && opts[0] != nil {
		timeout = opts[0].Timeout
		filter = opts[0].Filter
		reply = opts[0].Reply
	}

	if err := s.manager.SetStateWithOpts(s.Key(), step, s.raw.Payload, timeout, filter); err != nil {
		return err
	}

	if text == "" {
		return nil
	}

	if reply && s.ReplyFn != nil {
		return s.ReplyFn(text)
	}
	if s.SendFn != nil {
		return s.SendFn(text)
	}
	return nil
}

func (s *State) NextMedia(step string, media tg.InputMediaClass, caption string, opts ...*NextOpts) error {
	s.raw.Step = step
	if err := s.save(); err != nil {
		return err
	}

	var timeout time.Duration
	var filter Filter
	if len(opts) > 0 && opts[0] != nil {
		timeout = opts[0].Timeout
		filter = opts[0].Filter
	}

	if err := s.manager.SetStateWithOpts(s.Key(), step, s.raw.Payload, timeout, filter); err != nil {
		return err
	}
	if s.MediaFn != nil {
		return s.MediaFn(media, caption)
	}
	return nil
}

func (s *State) End(text string, opts ...*EndOpts) error {
	if err := s.manager.ClearState(s.Key()); err != nil {
		return err
	}
	if text == "" {
		return nil
	}

	var reply bool
	if len(opts) > 0 && opts[0] != nil {
		reply = opts[0].Reply
	}

	if reply && s.ReplyFn != nil {
		return s.ReplyFn(text)
	}
	if s.SendFn != nil {
		return s.SendFn(text)
	}
	return nil
}
