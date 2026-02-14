package filters

import (
	"bytes"
	"regexp"

	"github.com/gotd/td/tg"
)

type businessCallbackQuery struct{}

func (*businessCallbackQuery) All(_ *tg.UpdateBusinessBotCallbackQuery) bool {
	return true
}

func (*businessCallbackQuery) ConnectionID(id string) BusinessCallbackQueryFilter {
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return bq.ConnectionID == id
	}
}

func (*businessCallbackQuery) FromUserID(userID int64) BusinessCallbackQueryFilter {
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return bq.UserID == userID
	}
}

func (*businessCallbackQuery) Prefix(prefix string) BusinessCallbackQueryFilter {
	if len(prefix) == 0 {
		return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
			return len(bq.Data) > 0
		}
	}
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return bytes.HasPrefix(bq.Data, []byte(prefix))
	}
}

func (*businessCallbackQuery) Suffix(suffix string) BusinessCallbackQueryFilter {
	if len(suffix) == 0 {
		return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
			return len(bq.Data) > 0
		}
	}
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return bytes.HasSuffix(bq.Data, []byte(suffix))
	}
}

func (*businessCallbackQuery) Equal(data string) BusinessCallbackQueryFilter {
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return bytes.Equal(bq.Data, []byte(data))
	}
}

func (*businessCallbackQuery) Contains(substring string) BusinessCallbackQueryFilter {
	if len(substring) == 0 {
		return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
			return len(bq.Data) > 0
		}
	}
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return bytes.Contains(bq.Data, []byte(substring))
	}
}

func (*businessCallbackQuery) Regex(pattern string) BusinessCallbackQueryFilter {
	r := regexp.MustCompile(pattern)
	return func(bq *tg.UpdateBusinessBotCallbackQuery) bool {
		return r.Match(bq.Data)
	}
}
