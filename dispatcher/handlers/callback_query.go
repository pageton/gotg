package handlers

import (
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/adapter"
)

// CallbackQuery handler is executed when the update consists of tg.UpdateBotCallbackQuery.
type CallbackQuery struct {
	Filters       filters.CallbackQueryFilter
	Callback      CallbackResponse
	UpdateFilters filters.UpdateFilter
}

// NewCallbackQuery creates a new CallbackQuery handler bound to call its response.
func NewCallbackQuery(filters filters.CallbackQueryFilter, response CallbackResponse) CallbackQuery {
	return CallbackQuery{
		Filters:       filters,
		Callback:      response,
		UpdateFilters: nil,
	}
}

// OnCallbackQuery creates a new CallbackQuery handler with an UpdateHandler.
// This is a convenience function for handlers that only need to Update parameter.
func OnCallbackQuery(filter filters.CallbackQueryFilter, handler UpdateHandler, updateFilters ...filters.UpdateFilter) CallbackQuery {
	return CallbackQuery{
		Filters:  filter,
		Callback: ToCallbackResponse(handler),
		UpdateFilters: func(u *adapter.Update) bool {
			if len(updateFilters) == 0 {
				return true
			}
			for _, f := range updateFilters {
				if !f(u) {
					return false
				}
			}
			return true
		},
	}
}

func (c CallbackQuery) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.CallbackQuery == nil {
		return nil
	}
	if c.Filters != nil && !c.Filters(u.CallbackQuery) {
		return nil
	}
	if c.UpdateFilters != nil && !c.UpdateFilters(u) {
		return nil
	}
	return c.Callback(ctx, u)
}
