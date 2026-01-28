package filters

import "github.com/gotd/td/tg"

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
