package dispatcher

import (
	"sort"

	"github.com/pageton/gotg/adapter"
)

// Handler is the common interface for all the handlers.
type Handler interface {
	// CheckUpdate checks whether the update should be handled by this handler and processes it.
	CheckUpdate(*adapter.Context, *adapter.Update) error
}

// AddHandler adds a new handler to the dispatcher. The dispatcher will call CheckUpdate() to see whether the handler
// should be executed, and then execute it.
func (dp *NativeDispatcher) AddHandler(h Handler) {
	group := int(dp.nextGroup.Load())
	dp.nextGroup.Add(1)
	dp.AddHandlerToGroup(h, group)
}

// AddHandlerToGroup adds a handler to a specific group; lowest number will be processed first.
// Optimized: Uses tracking of sorted state to avoid unnecessary sorting.
func (dp *NativeDispatcher) AddHandlerToGroup(h Handler, group int) {
	dp.handlerGroupsMu.Lock()
	defer dp.handlerGroupsMu.Unlock()

	handlers, ok := dp.handlerMap[group]
	if !ok {
		dp.handlerGroups = append(dp.handlerGroups, group)
		// Optimized: Only sort if group wasn't present before
		// Check if already sorted (new group is at the end)
		if len(dp.handlerGroups) > 1 && dp.handlerGroups[len(dp.handlerGroups)-2] > group {
			sort.Ints(dp.handlerGroups)
		}
	}
	// Optimized: Pre-allocate slice with capacity if this is a new group
	if len(handlers) == 0 {
		handlers = make([]Handler, 0, 4) // Pre-allocate for 4 handlers
	}
	dp.handlerMap[group] = append(handlers, h)
}

// AddHandlers adds multiple handlers to the dispatcher with auto-assigned groups
func (dp *NativeDispatcher) AddHandlers(handlers ...Handler) {
	for _, h := range handlers {
		dp.AddHandler(h)
	}
}

// AddHandlersToGroup adds multiple handlers to a specific group
// Optimized: Pre-allocate handler slice for all handlers at once
func (dp *NativeDispatcher) AddHandlersToGroup(group int, handlers ...Handler) {
	if len(handlers) == 0 {
		return
	}

	dp.handlerGroupsMu.Lock()
	defer dp.handlerGroupsMu.Unlock()

	existing, ok := dp.handlerMap[group]
	if !ok {
		dp.handlerGroups = append(dp.handlerGroups, group)
		// Check if sorting is needed
		if len(dp.handlerGroups) > 1 && dp.handlerGroups[len(dp.handlerGroups)-2] > group {
			sort.Ints(dp.handlerGroups)
		}
		// Pre-allocate with exact capacity
		dp.handlerMap[group] = handlers
	} else {
		// Append to existing handlers
		dp.handlerMap[group] = append(existing, handlers...)
	}
}
