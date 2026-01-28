package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

// PendingJoinRequests handler is executed on all type of incoming updates.
type PendingJoinRequests struct {
	Callback CallbackResponse
	Filters  filters.PendingJoinRequestsFilter
}

// NewChatJoinRequest creates a new AnyUpdate handler bound to call its response.
func NewChatJoinRequest(filters filters.PendingJoinRequestsFilter, response CallbackResponse) PendingJoinRequests {
	return PendingJoinRequests{Callback: response, Filters: filters}
}

func OnChatJoinRequest(handler UpdateHandler, requestFilters ...filters.PendingJoinRequestsFilter) PendingJoinRequests {
	var filter filters.PendingJoinRequestsFilter
	if len(requestFilters) > 0 {
		filter = requestFilters[0]
	}
	return PendingJoinRequests{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (c PendingJoinRequests) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.ChatJoinRequest == nil {
		return nil
	}
	if c.Filters != nil && !c.Filters(u.ChatJoinRequest) {
		return nil
	}
	return c.Callback(ctx, u)
}
