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
func (dp *NativeDispatcher) AddHandlerToGroup(h Handler, group int) {
	handlers, ok := dp.handlerMap[group]
	if !ok {
		dp.handlerGroups = append(dp.handlerGroups, group)
		sort.Ints(dp.handlerGroups)
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
func (dp *NativeDispatcher) AddHandlersToGroup(group int, handlers ...Handler) {
	for _, h := range handlers {
		dp.AddHandlerToGroup(h, group)
	}
}
