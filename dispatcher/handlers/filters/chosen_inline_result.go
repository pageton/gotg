package filters

import "github.com/gotd/td/tg"

type chosenInlineResult struct{}

func (*chosenInlineResult) All(_ *tg.UpdateBotInlineSend) bool {
	return true
}

func (*chosenInlineResult) FromUserID(userID int64) ChosenInlineResultFilter {
	return func(cir *tg.UpdateBotInlineSend) bool {
		return cir.UserID == userID
	}
}

func (*chosenInlineResult) ResultID(id string) ChosenInlineResultFilter {
	return func(cir *tg.UpdateBotInlineSend) bool {
		return cir.ID == id
	}
}

func (*chosenInlineResult) QueryPrefix(prefix string) ChosenInlineResultFilter {
	return func(cir *tg.UpdateBotInlineSend) bool {
		return len(cir.Query) >= len(prefix) && cir.Query[:len(prefix)] == prefix
	}
}
