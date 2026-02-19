package middleware

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
)

// Chain creates a middleware chain that executes middlewares in order.
// Each middleware can decide to continue or stop the chain.
// The chain is safe for concurrent use — each CheckUpdate call
// gets its own index counter.
func Chain(middlewares ...dispatcher.Handler) dispatcher.Handler {
	return &chainHandler{middlewares: middlewares}
}

type chainHandler struct {
	middlewares []dispatcher.Handler
}

func (ch *chainHandler) CheckUpdate(ctx *adapter.Context, update *adapter.Update) error {
	return ch.run(ctx, update, 0)
}

func (ch *chainHandler) run(ctx *adapter.Context, update *adapter.Update, index int) error {
	if index >= len(ch.middlewares) {
		return nil
	}

	err := ch.middlewares[index].CheckUpdate(ctx, update)
	switch err {
	case dispatcher.StopClient, dispatcher.EndGroups:
		return err
	case dispatcher.SkipCurrentGroup:
		return ch.run(ctx, update, index+2)
	case dispatcher.ContinueGroups, nil:
		return ch.run(ctx, update, index+1)
	default:
		return err
	}
}
