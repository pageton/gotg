package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

// Message handler is executed when the update consists of tg.Message with provided conditions.
type Message struct {
	Callback      CallbackResponse
	Filters       filters.MessageFilter
	UpdateFilters filters.UpdateFilter
	Outgoing      bool
}

// NewMessage creates a new Message handler bound to call its response.
func NewMessage(filters filters.MessageFilter, response CallbackResponse) Message {
	return Message{
		Callback:      response,
		Filters:       filters,
		UpdateFilters: nil,
		Outgoing:      true,
	}
}

// OnMessage creates a new Message handler with an UpdateHandler.
// This is a convenience function for handlers that only need to Update parameter.
func OnMessage(handler UpdateHandler, messageFilters ...filters.MessageFilter) Message {
	var filter filters.MessageFilter
	if len(messageFilters) > 0 {
		filter = messageFilters[0]
		if len(messageFilters) > 1 {
			filter = filters.MessageAnd(messageFilters...)
		}
	}
	return Message{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
		Outgoing:      true,
	}
}

func (m Message) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.IsEdited || !u.HasMessage() {
		return nil
	}
	msg := u.EffectiveMessage
	if !m.Outgoing && msg.IsOutgoing() {
		return nil
	}
	if m.Filters != nil && !m.Filters(msg) {
		return nil
	}
	if m.UpdateFilters != nil && !m.UpdateFilters(u) {
		return nil
	}
	return m.Callback(ctx, u)
}
