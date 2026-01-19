package filters

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/types"
)

var (
	Message             = messageFilters{}
	CallbackQuery       = callbackQueryFilters{}
	InlineQuery         = inlineQuery{}
	PendingJoinRequests = pendingJoinRequests{}
	ChatMemberUpdated   = chatMemberUpdated{}
)

type (
	UpdateFilter              func(u *adapter.Update) bool
	MessageFilter             func(m *types.Message) bool
	CallbackQueryFilter       func(cbq *tg.UpdateBotCallbackQuery) bool
	InlineQueryFilter         func(iq *tg.UpdateBotInlineQuery) bool
	PendingJoinRequestsFilter func(cjr *tg.UpdatePendingJoinRequests) bool
	ChatMemberUpdatedFilter   func(u *adapter.Update) bool
)

type AndFilter struct {
	filters []UpdateFilter
}

type OrFilter struct {
	filters []UpdateFilter
}

type NotFilter struct {
	filter UpdateFilter
}

func (a AndFilter) Check(u *adapter.Update) bool {
	for _, f := range a.filters {
		if !f(u) {
			return false
		}
	}
	return true
}

func (o OrFilter) Check(u *adapter.Update) bool {
	for _, f := range o.filters {
		if f(u) {
			return true
		}
	}
	return false
}

func (n NotFilter) Check(u *adapter.Update) bool {
	return !n.filter(u)
}

// Private returns true if the update is from a private chat.
func Private(u *adapter.Update) bool {
	return u.GetUserChat() != nil
}

func Incoming(u *adapter.Update) bool {
	return !u.EffectiveMessage.Out
}

func (a AndFilter) Call(u *adapter.Update) bool {
	return a.Check(u)
}

func (o OrFilter) Call(u *adapter.Update) bool {
	return o.Check(u)
}

func (n NotFilter) Call(u *adapter.Update) bool {
	return n.Check(u)
}

type MessageAndFilter struct {
	filters []MessageFilter
}

type MessageOrFilter struct {
	filters []MessageFilter
}

type MessageNotFilter struct {
	filter MessageFilter
}

func (a MessageAndFilter) Check(m *types.Message) bool {
	for _, f := range a.filters {
		if !f(m) {
			return false
		}
	}
	return true
}

func (o MessageOrFilter) Check(m *types.Message) bool {
	for _, f := range o.filters {
		if f(m) {
			return true
		}
	}
	return false
}

func (n MessageNotFilter) Check(m *types.Message) bool {
	return !n.filter(m)
}

func (a MessageAndFilter) Call(m *types.Message) bool {
	return a.Check(m)
}

func (o MessageOrFilter) Call(m *types.Message) bool {
	return o.Check(m)
}

func (n MessageNotFilter) Call(m *types.Message) bool {
	return n.Check(m)
}

func And(filters ...UpdateFilter) UpdateFilter {
	return func(u *adapter.Update) bool {
		for _, f := range filters {
			if !f(u) {
				return false
			}
		}
		return true
	}
}

func Or(filters ...UpdateFilter) UpdateFilter {
	return func(u *adapter.Update) bool {
		for _, f := range filters {
			if f(u) {
				return true
			}
		}
		return false
	}
}

func Not(filter UpdateFilter) UpdateFilter {
	return func(u *adapter.Update) bool {
		return !filter(u)
	}
}

func MessageAnd(filters ...MessageFilter) MessageFilter {
	return func(m *types.Message) bool {
		for _, f := range filters {
			if !f(m) {
				return false
			}
		}
		return true
	}
}

func MessageOr(filters ...MessageFilter) MessageFilter {
	return func(m *types.Message) bool {
		for _, f := range filters {
			if f(m) {
				return true
			}
		}
		return false
	}
}

func MessageNot(filter MessageFilter) MessageFilter {
	return func(m *types.Message) bool {
		return !filter(m)
	}
}

// Supergroup returns true if the update is from a supergroup.
func Supergroup(u *adapter.Update) bool {
	if c := u.GetChannel(); c != nil {
		return c.Megagroup
	}
	return false
}

// Channel returns true if the update is from a channel.
func Channel(u *adapter.Update) bool {
	channelType := u.GetChannel()
	chatType := u.GetChat()
	if channelType != nil && chatType == nil {
		return !channelType.Megagroup
	}
	return false
}

// Group returns true if the update is from a normal group.
func Group(u *adapter.Update) bool {
	return u.GetChat() != nil
}
