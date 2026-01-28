package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type DeletedMessage struct {
	Callback CallbackResponse
	Filters  filters.DeletedMessageFilter
}

func NewDeletedMessage(filters filters.DeletedMessageFilter, response CallbackResponse) DeletedMessage {
	return DeletedMessage{
		Callback: response,
		Filters:  filters,
	}
}

func OnDeletedMessage(handler UpdateHandler, deletedFilters ...filters.DeletedMessageFilter) DeletedMessage {
	var filter filters.DeletedMessageFilter
	if len(deletedFilters) > 0 {
		filter = deletedFilters[0]
	}
	return DeletedMessage{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (d DeletedMessage) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.DeletedMessages == nil && u.DeletedChannelMessages == nil {
		return nil
	}
	if d.Filters != nil && !d.Filters(u) {
		return nil
	}
	return d.Callback(ctx, u)
}
