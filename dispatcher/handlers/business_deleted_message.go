package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type BusinessDeletedMessage struct {
	Callback CallbackResponse
	Filters  filters.BusinessDeletedMessageFilter
}

func NewBusinessDeletedMessage(f filters.BusinessDeletedMessageFilter, response CallbackResponse) BusinessDeletedMessage {
	return BusinessDeletedMessage{
		Callback: response,
		Filters:  f,
	}
}

func OnBusinessDeletedMessage(handler UpdateHandler, f ...filters.BusinessDeletedMessageFilter) BusinessDeletedMessage {
	var filter filters.BusinessDeletedMessageFilter
	if len(f) > 0 {
		filter = f[0]
	}
	return BusinessDeletedMessage{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (h BusinessDeletedMessage) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.BusinessDeletedMessages == nil {
		return nil
	}
	if h.Filters != nil && !h.Filters(u.BusinessDeletedMessages) {
		return nil
	}
	return h.Callback(ctx, u)
}
