package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type BusinessCallbackQuery struct {
	Callback CallbackResponse
	Filters  filters.BusinessCallbackQueryFilter
}

func NewBusinessCallbackQuery(f filters.BusinessCallbackQueryFilter, response CallbackResponse) BusinessCallbackQuery {
	return BusinessCallbackQuery{
		Callback: response,
		Filters:  f,
	}
}

func OnBusinessCallbackQuery(handler UpdateHandler, f ...filters.BusinessCallbackQueryFilter) BusinessCallbackQuery {
	var filter filters.BusinessCallbackQueryFilter
	if len(f) > 0 {
		filter = f[0]
	}
	return BusinessCallbackQuery{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (h BusinessCallbackQuery) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.BusinessCallbackQuery == nil {
		return nil
	}
	if h.Filters != nil && !h.Filters(u.BusinessCallbackQuery) {
		return nil
	}
	return h.Callback(ctx, u)
}
