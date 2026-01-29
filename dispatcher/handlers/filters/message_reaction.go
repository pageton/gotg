package filters

import "github.com/gotd/td/tg"

type messageReaction struct{}

func (*messageReaction) All(_ *tg.UpdateBotMessageReaction) bool {
	return true
}

func (*messageReaction) FromMsgID(msgID int) MessageReactionFilter {
	return func(mr *tg.UpdateBotMessageReaction) bool {
		return mr.MsgID == msgID
	}
}
