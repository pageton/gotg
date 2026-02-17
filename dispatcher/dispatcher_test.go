package dispatcher

import (
	"testing"

	"github.com/pageton/gotg/adapter"
)

type mockHandler struct {
	retErr error
	called int
}

func (m *mockHandler) CheckUpdate(_ *adapter.Context, _ *adapter.Update) error {
	m.called++
	return m.retErr
}

func newTestDispatcher() *NativeDispatcher {
	nd := &NativeDispatcher{
		handlerMap:    make(map[int][]Handler),
		handlerGroups: make([]int, 0, 8),
		updateSem:     make(chan struct{}, 10),
	}
	return nd
}

func TestAddHandler_AutoIncrements(t *testing.T) {
	dp := newTestDispatcher()

	h1 := &mockHandler{}
	h2 := &mockHandler{}
	dp.AddHandler(h1)
	dp.AddHandler(h2)

	if len(dp.handlerGroups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(dp.handlerGroups))
	}
	if dp.handlerGroups[0] != 0 || dp.handlerGroups[1] != 1 {
		t.Fatalf("expected groups [0, 1], got %v", dp.handlerGroups)
	}
}

func TestAddHandlerToGroup(t *testing.T) {
	dp := newTestDispatcher()

	h1 := &mockHandler{}
	h2 := &mockHandler{}
	dp.AddHandlerToGroup(h1, 5)
	dp.AddHandlerToGroup(h2, 5)

	if len(dp.handlerGroups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(dp.handlerGroups))
	}
	if len(dp.handlerMap[5]) != 2 {
		t.Fatalf("expected 2 handlers in group 5, got %d", len(dp.handlerMap[5]))
	}
}

func TestAddHandlersToGroup(t *testing.T) {
	dp := newTestDispatcher()

	h1 := &mockHandler{}
	h2 := &mockHandler{}
	h3 := &mockHandler{}
	dp.AddHandlersToGroup(10, h1, h2, h3)

	if len(dp.handlerMap[10]) != 3 {
		t.Fatalf("expected 3 handlers, got %d", len(dp.handlerMap[10]))
	}
}

func TestHandlerGroupOrdering(t *testing.T) {
	dp := newTestDispatcher()

	dp.AddHandlerToGroup(&mockHandler{}, 10)
	dp.AddHandlerToGroup(&mockHandler{}, 1)
	dp.AddHandlerToGroup(&mockHandler{}, 5)

	expected := []int{1, 5, 10}
	if len(dp.handlerGroups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(dp.handlerGroups))
	}
	for i, g := range dp.handlerGroups {
		if g != expected[i] {
			t.Fatalf("expected group %d at index %d, got %d", expected[i], i, g)
		}
	}
}

func TestSentinelErrors(t *testing.T) {
	if ContinueGroups.Error() != "continued" {
		t.Fatalf("unexpected ContinueGroups message: %s", ContinueGroups.Error())
	}
	if EndGroups.Error() != "stopped" {
		t.Fatalf("unexpected EndGroups message: %s", EndGroups.Error())
	}
	if SkipCurrentGroup.Error() != "skipped" {
		t.Fatalf("unexpected SkipCurrentGroup message: %s", SkipCurrentGroup.Error())
	}
	if StopClient.Error() != "disconnect" {
		t.Fatalf("unexpected StopClient message: %s", StopClient.Error())
	}
}

func TestAddHandlersToGroup_Empty(t *testing.T) {
	dp := newTestDispatcher()
	dp.AddHandlersToGroup(1)

	if len(dp.handlerGroups) != 0 {
		t.Fatalf("expected 0 groups for empty AddHandlersToGroup, got %d", len(dp.handlerGroups))
	}
}

func TestAddHandlersToGroup_AppendToExisting(t *testing.T) {
	dp := newTestDispatcher()

	h1 := &mockHandler{}
	h2 := &mockHandler{}
	dp.AddHandlerToGroup(h1, 3)
	dp.AddHandlersToGroup(3, h2)

	if len(dp.handlerMap[3]) != 2 {
		t.Fatalf("expected 2 handlers in group 3, got %d", len(dp.handlerMap[3]))
	}
}
