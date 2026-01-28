package filters

import "github.com/gotd/td/tg"

type chatBoost struct{}

func (*chatBoost) All(_ *tg.UpdateBotChatBoost) bool {
	return true
}
