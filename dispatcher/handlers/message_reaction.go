package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

type MessageReaction struct {
	Callback CallbackResponse
	Filters  filters.MessageReactionFilter
}

func NewMessageReaction(filters filters.MessageReactionFilter, response CallbackResponse) MessageReaction {
	return MessageReaction{
		Callback: response,
		Filters:  filters,
	}
}

func OnMessageReaction(handler UpdateHandler, reactionFilters ...filters.MessageReactionFilter) MessageReaction {
	var filter filters.MessageReactionFilter
	if len(reactionFilters) > 0 {
		filter = reactionFilters[0]
	}
	return MessageReaction{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (mr MessageReaction) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.MessageReaction == nil {
		return nil
	}
	if mr.Filters != nil && !mr.Filters(u.MessageReaction) {
		return nil
	}
	return mr.Callback(ctx, u)
}
