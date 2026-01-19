package filters

import (
	"strings"

	"github.com/gotd/td/tg"
)

type callbackQueryFilters struct{}

// All returns true on every type of tg.UpdateBotCallbackQuery update.
func (*callbackQueryFilters) All(_ *tg.UpdateBotCallbackQuery) bool {
	return true
}

// Prefix returns true if the tg.UpdateBotCallbackQuery's Data field contains provided prefix.
func (*callbackQueryFilters) Prefix(prefix string) CallbackQueryFilter {
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return strings.HasPrefix(string(cbq.Data), prefix)
	}
}

// Suffix returns true if the tg.UpdateBotCallbackQuery's Data field contains provided suffix.
func (*callbackQueryFilters) Suffix(suffix string) CallbackQueryFilter {
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return strings.HasSuffix(string(cbq.Data), suffix)
	}
}

// Equal checks if the tg.UpdateBotCallbackQuery's Data field is equal to the provided data and returns true if matches.
func (*callbackQueryFilters) Equal(data string) CallbackQueryFilter {
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return string(cbq.Data) == data
	}
}

// FromUserID checks if the tg.UpdateBotCallbackQuery was sent by the provided user id and returns true if matches.
func (*callbackQueryFilters) FromUserID(userID int64) CallbackQueryFilter {
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return cbq.UserID == userID
	}
}

// GameName checks if the tg.UpdateBotCallbackQuery's GameShortName field is equal to the provided name and returns true if matches.
func (*callbackQueryFilters) GameName(name string) CallbackQueryFilter {
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return cbq.GameShortName == name
	}
}
