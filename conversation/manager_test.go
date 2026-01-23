package conversation

import (
	"fmt"
	"testing"
	"time"

	"github.com/gotd/td/tg"
	mtp_errors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
	"gorm.io/driver/sqlite"
)

func TestManager_WaitResponse(t *testing.T) {
	mgr := NewManager(nil, 200*time.Millisecond)
	key := Key{ChatID: 1, UserID: 2}
	ctx := t.Context()

	done := make(chan error, 1)
	go func() {
		msg, err := mgr.WaitResponse(ctx, key, Options{})
		if err != nil {
			done <- err
			return
		}
		if msg.ID != 10 {
			done <- fmt.Errorf("unexpected message id %d", msg.ID)
			return
		}
		done <- nil
	}()

	time.Sleep(20 * time.Millisecond)
	msg := &types.Message{Message: &tg.Message{ID: 10}}
	if !mgr.Route(key, msg) {
		t.Fatalf("route returned false")
	}

	if err := <-done; err != nil {
		t.Fatalf("waiter failed: %v", err)
	}
}

func TestManager_Timeout(t *testing.T) {
	mgr := NewManager(nil, 50*time.Millisecond)
	key := Key{ChatID: 3, UserID: 4}
	ctx := t.Context()

	start := time.Now()
	_, err := mgr.WaitResponse(ctx, key, Options{})
	if err == nil || err != mtp_errors.ErrConversationTimeout {
		t.Fatalf("expected timeout, got %v", err)
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatalf("timeout fired too early")
	}
}

func TestManager_PersistenceHooks(t *testing.T) {
	ps := storage.NewPeerStorage(sqlite.Open("file::memory:?cache=shared"), false)
	if ps == nil {
		t.Fatalf("nil peer storage")
	}
	mgr := NewManager(ps, 500*time.Millisecond)
	key := Key{ChatID: 11, UserID: 99}
	ctx := t.Context()

	done := make(chan error, 1)
	go func() {
		_, err := mgr.WaitResponse(ctx, key, Options{Step: "collect_name", Timeout: 200 * time.Millisecond})
		done <- err
	}()

	time.Sleep(20 * time.Millisecond)
	stateKey := storage.ConversationKey(key.ChatID, key.UserID)
	var state storage.ConversationState
	if err := ps.SqlSession.Where("key = ?", stateKey).First(&state).Error; err != nil {
		t.Fatalf("expected state persisted: %v", err)
	}
	if state.Step != "collect_name" {
		t.Fatalf("unexpected step %s", state.Step)
	}

	mgr.Route(key, &types.Message{Message: &tg.Message{ID: 7}})
	if err := <-done; err != nil {
		t.Fatalf("waiter ended with err %v", err)
	}

	if err := ps.SqlSession.Where("key = ?", stateKey).First(&state).Error; err == nil {
		t.Fatalf("state should have been removed")
	}
}
