package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type OutgoingMessage struct {
	Callback      CallbackResponse
	Filters       filters.MessageFilter
	UpdateFilters filters.UpdateFilter
}

func NewOutgoing(filters filters.MessageFilter, response CallbackResponse) OutgoingMessage {
	return OutgoingMessage{
		Callback:      response,
		Filters:       filters,
		UpdateFilters: nil,
	}
}

func OnOutgoing(handler UpdateHandler, messageFilters ...filters.MessageFilter) OutgoingMessage {
	var filter filters.MessageFilter
	if len(messageFilters) > 0 {
		filter = messageFilters[0]
		if len(messageFilters) > 1 {
			filter = filters.MessageAnd(messageFilters...)
		}
	}
	return OutgoingMessage{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
	}
}

func (m OutgoingMessage) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.EffectiveOutgoing == nil {
		return nil
	}
	if m.Filters != nil && !m.Filters(u.EffectiveMessage) {
		return nil
	}
	if m.UpdateFilters != nil && !m.UpdateFilters(u) {
		return nil
	}
	return m.Callback(ctx, u)
}
