package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

// InlineQuery handler is executed when the update consists of tg.UpdateBotInlineQuery.
type InlineQuery struct {
	Callback      CallbackResponse
	Filters       filters.InlineQueryFilter
	UpdateFilters filters.UpdateFilter
}

// NewInlineQuery creates a new InlineQuery handler bound to call its response.
func NewInlineQuery(filters filters.InlineQueryFilter, response CallbackResponse) InlineQuery {
	return InlineQuery{
		Filters:       filters,
		Callback:      response,
		UpdateFilters: nil,
	}
}

func OnInlineQuery(handler UpdateHandler, inlineFilters ...filters.InlineQueryFilter) InlineQuery {
	var filter filters.InlineQueryFilter
	if len(inlineFilters) > 0 {
		filter = inlineFilters[0]
	}
	return InlineQuery{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
	}
}

func (c InlineQuery) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.InlineQuery == nil {
		return nil
	}
	if c.Filters != nil && !c.Filters(u.InlineQuery.Raw()) {
		return nil
	}
	if c.UpdateFilters != nil && !c.UpdateFilters(u) {
		return nil
	}
	return c.Callback(ctx, u)
}
