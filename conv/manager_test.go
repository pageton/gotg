package conv

import (
	"testing"
	"time"

	"github.com/pageton/gotg/storage"
)

func TestManager_RegisterStep(t *testing.T) {
	m := NewManager(nil, 30*time.Second)

	m.RegisterStep("test_step", func(state *State) error {
		return nil
	})

	handler, ok := m.GetStepHandler("test_step")
	if !ok {
		t.Fatal("expected handler to be registered")
	}

	if handler == nil {
		t.Fatal("expected handler to not be nil")
	}

	_, notFound := m.GetStepHandler("nonexistent")
	if notFound {
		t.Fatal("expected handler to not be found")
	}
}

func TestManager_SetState_GetState_ClearState(t *testing.T) {
	m := NewManager(nil, 30*time.Second)

	key := Key{ChatID: 123, UserID: 456}

	err := m.SetState(key, "step1", []byte(`{"foo":"bar"}`))
	if err != nil {
		t.Fatalf("SetState failed: %v", err)
	}

	state, err := m.GetState(key)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if state != nil {
		t.Fatal("expected state to be nil with nil storage")
	}

	err = m.ClearState(key)
	if err != nil {
		t.Fatalf("ClearState failed: %v", err)
	}
}

func TestState_GetSet(t *testing.T) {
	raw := &storage.ConvState{
		Key:     "123:456",
		ChatID:  123,
		UserID:  456,
		Step:    "test",
		Payload: []byte(`{"name":"John","count":42}`),
	}

	ps, _ := storage.NewPeerStorage(nil, true)
	m := NewManager(ps, 30*time.Second)
	state := newState(raw, m)

	if state.GetString("name") != "John" {
		t.Fatalf("expected John, got %s", state.GetString("name"))
	}

	if state.GetInt("count") != 42 {
		t.Fatalf("expected 42, got %d", state.GetInt("count"))
	}

	state.Set("age", 25)
	if state.GetInt("age") != 25 {
		t.Fatalf("expected 25, got %d", state.GetInt("age"))
	}

	state.Delete("age")
	if state.Get("age") != nil {
		t.Fatal("expected age to be deleted")
	}
}

func TestState_Key(t *testing.T) {
	raw := &storage.ConvState{
		Key:    "123:456",
		ChatID: 123,
		UserID: 456,
		Step:   "test",
	}

	ps, _ := storage.NewPeerStorage(nil, true)
	m := NewManager(ps, 30*time.Second)
	state := newState(raw, m)

	key := state.Key()
	if key.ChatID != 123 || key.UserID != 456 {
		t.Fatalf("expected 123:456, got %d:%d", key.ChatID, key.UserID)
	}
}

func TestKey_String(t *testing.T) {
	key := Key{ChatID: 100, UserID: 200}
	if key.String() != "100:200" {
		t.Fatalf("expected 100:200, got %s", key.String())
	}
}
