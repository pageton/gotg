package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type BusinessEditedMessage struct {
	Callback      CallbackResponse
	Filters       filters.MessageFilter
	UpdateFilters filters.UpdateFilter
}

func NewBusinessEditedMessage(f filters.MessageFilter, response CallbackResponse) BusinessEditedMessage {
	return BusinessEditedMessage{
		Callback:      response,
		Filters:       f,
		UpdateFilters: nil,
	}
}

func OnBusinessEditedMessage(handler UpdateHandler, messageFilters ...filters.MessageFilter) BusinessEditedMessage {
	var filter filters.MessageFilter
	if len(messageFilters) > 0 {
		filter = messageFilters[0]
		if len(messageFilters) > 1 {
			filter = filters.MessageAnd(messageFilters...)
		}
	}
	return BusinessEditedMessage{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
	}
}

func (h BusinessEditedMessage) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.BusinessEditedMessage == nil {
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
