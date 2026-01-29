package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type ChosenInlineResult struct {
	Callback      CallbackResponse
	Filters       filters.ChosenInlineResultFilter
	UpdateFilters filters.UpdateFilter
}

func NewChosenInlineResult(filters filters.ChosenInlineResultFilter, response CallbackResponse) ChosenInlineResult {
	return ChosenInlineResult{
		Filters:       filters,
		Callback:      response,
		UpdateFilters: nil,
	}
}

func OnChosenInlineResult(handler UpdateHandler, resultFilters ...filters.ChosenInlineResultFilter) ChosenInlineResult {
	var filter filters.ChosenInlineResultFilter
	if len(resultFilters) > 0 {
		filter = resultFilters[0]
	}
	return ChosenInlineResult{
		Callback:      ToCallbackResponse(handler),
		Filters:       filter,
		UpdateFilters: nil,
	}
}

func (c ChosenInlineResult) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.ChosenInlineResult == nil {
		return nil
	}
	if c.Filters != nil && !c.Filters(u.ChosenInlineResult) {
		return nil
	}
	if c.UpdateFilters != nil && !c.UpdateFilters(u) {
		return nil
	}
	return c.Callback(ctx, u)
}
