package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type BusinessMessage struct {
	Callback      CallbackResponse
	Filters       filters.MessageFilter
	UpdateFilters filters.UpdateFilter
}

func NewBusinessMessage(f filters.MessageFilter, response CallbackResponse) BusinessMessage {
	return BusinessMessage{
		Callback:      response,
		Filters:       f,
		UpdateFilters: nil,
	}
}

func OnBusinessMessage(handler UpdateHandler, messageFilters ...filters.MessageFilter) BusinessMessage {
	var filter filters.MessageFilter
	if len(messageFilters) > 0 {
		filter = messageFilters[0]
		if len(messageFilters) > 1 {
			filter = filters.MessageAnd(messageFilters...)
		}
	}
	return BusinessMessage{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
	}
}

func (h BusinessMessage) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.BusinessMessage == nil {
		return nil
	}
	if !u.HasMessage() {
		return nil
	}
	msg := u.EffectiveMessage
	if h.Filters != nil && !h.Filters(msg) {
		return nil
	}
	if h.UpdateFilters != nil && !h.UpdateFilters(u) {
		return nil
	}
	return h.Callback(ctx, u)
}
