package filters

import (
	"bytes"

	"github.com/gotd/td/tg"
)

type callbackQueryFilters struct{}

// All returns true on every type of tg.UpdateBotCallbackQuery update.
func (*callbackQueryFilters) All(_ *tg.UpdateBotCallbackQuery) bool {
	return true
}

// Prefix returns true if the tg.UpdateBotCallbackQuery's Data field contains provided prefix.
// Optimized: Uses bytes.Index instead of string conversion
func (*callbackQueryFilters) Prefix(prefix string) CallbackQueryFilter {
	if len(prefix) == 0 {
		return func(cbq *tg.UpdateBotCallbackQuery) bool {
			return len(cbq.Data) > 0
		}
	}
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return bytes.HasPrefix(cbq.Data, []byte(prefix))
	}
}

// Suffix returns true if the tg.UpdateBotCallbackQuery's Data field contains provided suffix.
// Optimized: Uses bytes operations instead of string conversion
func (*callbackQueryFilters) Suffix(suffix string) CallbackQueryFilter {
	if len(suffix) == 0 {
		return func(cbq *tg.UpdateBotCallbackQuery) bool {
			return len(cbq.Data) > 0
		}
	}
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return bytes.HasSuffix(cbq.Data, []byte(suffix))
	}
}

// Equal checks if the tg.UpdateBotCallbackQuery's Data field is equal to the provided data and returns true if matches.
// Optimized: Uses bytes.Equal instead of string conversion
func (*callbackQueryFilters) Equal(data string) CallbackQueryFilter {
	return func(cbq *tg.UpdateBotCallbackQuery) bool {
		return bytes.Equal(cbq.Data, []byte(data))
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
