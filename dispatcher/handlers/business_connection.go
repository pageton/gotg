package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type BusinessConnection struct {
	Callback CallbackResponse
	Filters  filters.BusinessConnectionFilter
}

func NewBusinessConnection(f filters.BusinessConnectionFilter, response CallbackResponse) BusinessConnection {
	return BusinessConnection{
		Callback: response,
		Filters:  f,
	}
}

func OnBusinessConnection(handler UpdateHandler, f ...filters.BusinessConnectionFilter) BusinessConnection {
	var filter filters.BusinessConnectionFilter
	if len(f) > 0 {
		filter = f[0]
	}
	return BusinessConnection{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (h BusinessConnection) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.BusinessConnection == nil {
		return nil
	}
	if h.Filters != nil && !h.Filters(u.BusinessConnection) {
		return nil
	}
	return h.Callback(ctx, u)
}
