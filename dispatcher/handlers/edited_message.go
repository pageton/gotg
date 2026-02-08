package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type EditedMessage struct {
	Callback      CallbackResponse
	Filters       filters.MessageFilter
	UpdateFilters filters.UpdateFilter
	Outgoing      bool
}

func NewEditedMessage(filters filters.MessageFilter, response CallbackResponse) EditedMessage {
	return EditedMessage{
		Callback:      response,
		Filters:       filters,
		UpdateFilters: nil,
		Outgoing:      true,
	}
}

func OnEditedMessage(handler UpdateHandler, messageFilters ...filters.MessageFilter) EditedMessage {
	var filter filters.MessageFilter
	if len(messageFilters) > 0 {
		filter = messageFilters[0]
		if len(messageFilters) > 1 {
			filter = filters.MessageAnd(messageFilters...)
		}
	}
	return EditedMessage{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
		Outgoing:      true,
	}
}

func (m EditedMessage) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if !u.IsEdited || u.MessageReaction != nil {
		return nil
	}
	if !u.HasMessage() {
		return nil
	}
	msg := u.EffectiveMessage
	if msg.EditHide {
		return nil
	}
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
