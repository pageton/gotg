package handlers

import (
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
)

// ChatMemberUpdated handler is executed on all type of incoming updates.
type ChatMemberUpdated struct {
	Callback CallbackResponse
	Filters  filters.ChatMemberUpdatedFilter
}

// NewChatMemberUpdated creates a new ChatMemberUpdated handler bound to call its response.
func NewChatMemberUpdated(filters filters.ChatMemberUpdatedFilter, response CallbackResponse) ChatMemberUpdated {
	return ChatMemberUpdated{
		Callback: response,
		Filters:  filters,
	}
}

func NewParticipant(filters filters.ChatMemberUpdatedFilter, response CallbackResponse) ChatMemberUpdated {
	return NewChatMemberUpdated(filters, response)
}

func OnParticipant(handler UpdateHandler, chatMemberFilters ...filters.ChatMemberUpdatedFilter) ChatMemberUpdated {
	return OnChatMemberUpdated(handler, chatMemberFilters...)
}

func OnChatMemberUpdated(handler UpdateHandler, chatMemberFilters ...filters.ChatMemberUpdatedFilter) ChatMemberUpdated {
	var filter filters.ChatMemberUpdatedFilter
	if len(chatMemberFilters) > 0 {
		filter = chatMemberFilters[0]
	}
	return ChatMemberUpdated{
		Callback: ToCallbackResponse(handler),
		Filters:  filter,
	}
}

func (cm ChatMemberUpdated) CheckUpdate(ctx *adapter.Context, u *adapter.Update) error {
	if u.ChatParticipant == nil && u.ChannelParticipant == nil {
		return nil
	}
	if cm.Filters != nil && !cm.Filters(u) {
		return nil
	}
	return cm.Callback(ctx, u)
}
