package handlers

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

// CallbackQuery handler is executed when the update consists of tg.UpdateBotCallbackQuery
// or tg.UpdateInlineBotCallbackQuery (for inline message buttons).
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
	var cbqForFilter *tg.UpdateBotCallbackQuery

	switch {
	case u.CallbackQuery != nil:
		cbqForFilter = u.CallbackQuery
	case u.InlineCallbackQuery != nil:
		cbqForFilter = &tg.UpdateBotCallbackQuery{
			QueryID:       u.InlineCallbackQuery.QueryID,
			UserID:        u.InlineCallbackQuery.UserID,
			Data:          u.InlineCallbackQuery.Data,
			GameShortName: u.InlineCallbackQuery.GameShortName,
		}
	default:
		return nil
	}

	if c.Filters != nil && !c.Filters(cbqForFilter) {
		return nil
	}
	if c.UpdateFilters != nil && !c.UpdateFilters(u) {
		return nil
	}
	return c.Callback(ctx, u)
}
