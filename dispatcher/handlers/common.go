package handlers

import (
	"github.com/pageton/gotg/adapter"
)

// CallbackResponse is the function which will be called on a handler's execution.
type CallbackResponse func(*adapter.Context, *adapter.Update) error

// UpdateHandler is the function signature for handlers that only need Update.
// The Context can be accessed via update.Ctx if needed.
type UpdateHandler func(*adapter.Update) error

// ToCallbackResponse converts an UpdateHandler to CallbackResponse.
func ToCallbackResponse(h UpdateHandler) CallbackResponse {
	return func(_ *adapter.Context, u *adapter.Update) error {
		return h(u)
	}
}
