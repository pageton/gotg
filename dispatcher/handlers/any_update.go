package handlers

import (
	"github.com/pageton/gotg/adapter"
)

// AnyUpdate handler is executed on all type of incoming updates.
type AnyUpdate struct {
	Callback CallbackResponse
}

// NewAnyUpdate creates a new AnyUpdate handler bound to call its response.
func NewAnyUpdate(response CallbackResponse) AnyUpdate {
	return AnyUpdate{Callback: response}
}

// OnUpdate creates a new AnyUpdate handler with an UpdateHandler.
// This is a convenience function for handlers that only need the Update parameter.
func OnUpdate(handler UpdateHandler) AnyUpdate {
	return AnyUpdate{Callback: ToCallbackResponse(handler)}
}

func (au AnyUpdate) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	return au.Callback(ctx, u)
}
