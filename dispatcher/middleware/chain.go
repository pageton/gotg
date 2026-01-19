package middleware

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
)

// Chain creates a middleware chain that executes middlewares in order
// Each middleware can decide to continue or stop the chain
//
// Example:
//
//	dp.AddHandler(Chain(
//	    LoggingMiddleware(),
//	    RateLimitMiddleware(&RateLimitConfig{...}),
//	    I18nMiddleware(&I18nConfig{...}),
//	))
func Chain(middlewares ...dispatcher.Handler) dispatcher.Handler {
	return &chainHandler{
		middlewares: middlewares,
		index:       0,
	}
}

type chainHandler struct {
	middlewares []dispatcher.Handler
	index       int
}

func (ch *chainHandler) CheckUpdate(ctx *adapter.Context, update *adapter.Update) error {
	if ch.index >= len(ch.middlewares) {
		return nil
	}

	handler := ch.middlewares[ch.index]
	ch.index++

	err := handler.CheckUpdate(ctx, update)
	if err == dispatcher.StopClient || err == dispatcher.EndGroups {
		return err
	}
	if err == dispatcher.SkipCurrentGroup {
		ch.index++
		return ch.CheckUpdate(ctx, update)
	}
	if err == dispatcher.ContinueGroups {
		return ch.CheckUpdate(ctx, update)
	}
	return err
}
