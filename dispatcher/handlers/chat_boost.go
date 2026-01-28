package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type ChatBoost struct {
	Callback CallbackResponse
	Filters  filters.ChatBoostFilter
}

func NewChatBoost(filters filters.ChatBoostFilter, response CallbackResponse) ChatBoost {
	return ChatBoost{
		Callback: response,
		Filters:  filters,
	}
}

func OnChatBoost(handler UpdateHandler, boostFilters ...filters.ChatBoostFilter) ChatBoost {
	var filter filters.ChatBoostFilter
	if len(boostFilters) > 0 {
		filter = boostFilters[0]
	}
	return ChatBoost{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (cb ChatBoost) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.ChatBoost == nil {
		return nil
	}
	if cb.Filters != nil && !cb.Filters(u.ChatBoost) {
		return nil
	}
	return cb.Callback(ctx, u)
}
