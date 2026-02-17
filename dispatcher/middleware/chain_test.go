package middleware

import (
	"sync"
	"testing"

	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
)

type orderHandler struct {
	id    int
	order *[]int
	ret   error
}

func (h *orderHandler) CheckUpdate(_ *adapter.Context, _ *adapter.Update) error {
	*h.order = append(*h.order, h.id)
	return h.ret
}

func TestChain_ExecutesInOrder(t *testing.T) {
	var order []int
	chain := Chain(
		&orderHandler{id: 1, order: &order, ret: nil},
		&orderHandler{id: 2, order: &order, ret: nil},
		&orderHandler{id: 3, order: &order, ret: nil},
	)

	err := chain.CheckUpdate(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(order))
	}
	for i, v := range order {
		if v != i+1 {
			t.Fatalf("expected order %d at index %d, got %d", i+1, i, v)
		}
	}
}

func TestChain_EndGroupsStops(t *testing.T) {
	var order []int
	chain := Chain(
		&orderHandler{id: 1, order: &order, ret: nil},
		&orderHandler{id: 2, order: &order, ret: dispatcher.EndGroups},
		&orderHandler{id: 3, order: &order, ret: nil},
	)

	err := chain.CheckUpdate(nil, nil)
	if err != dispatcher.EndGroups {
		t.Fatalf("expected EndGroups, got %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(order))
	}
}

func TestChain_ContinueGroupsProceedsNormally(t *testing.T) {
	var order []int
	chain := Chain(
		&orderHandler{id: 1, order: &order, ret: dispatcher.ContinueGroups},
		&orderHandler{id: 2, order: &order, ret: nil},
	)

	err := chain.CheckUpdate(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(order))
	}
}

func TestChain_SkipCurrentGroupSkipsNext(t *testing.T) {
	var order []int
	chain := Chain(
		&orderHandler{id: 1, order: &order, ret: dispatcher.SkipCurrentGroup},
		&orderHandler{id: 2, order: &order, ret: nil},
		&orderHandler{id: 3, order: &order, ret: nil},
	)

	err := chain.CheckUpdate(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// SkipCurrentGroup at index 0 skips index 1, runs index 2
	if len(order) != 2 {
		t.Fatalf("expected 2 calls, got %d (order: %v)", len(order), order)
	}
	if order[0] != 1 || order[1] != 3 {
		t.Fatalf("expected [1, 3], got %v", order)
	}
}

func TestChain_ConcurrentSafe(t *testing.T) {
	chain := Chain(
		&orderHandler{id: 1, order: &[]int{}, ret: nil},
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = chain.CheckUpdate(nil, nil)
		}()
	}
	wg.Wait()
}

func TestChain_Empty(t *testing.T) {
	chain := Chain()
	err := chain.CheckUpdate(nil, nil)
	if err != nil {
		t.Fatalf("empty chain should return nil, got %v", err)
	}
}
