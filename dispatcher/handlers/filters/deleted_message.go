package filters

import "github.com/pageton/gotg/adapter"

type deletedMessage struct{}

func (*deletedMessage) All(_ *adapter.Update) bool {
	return true
}

func (*deletedMessage) ChannelOnly(u *adapter.Update) bool {
	return u.DeletedChannelMessages != nil
}

func (*deletedMessage) PrivateOnly(u *adapter.Update) bool {
	return u.DeletedMessages != nil
}

func (*deletedMessage) ChannelID(channelID int64) DeletedMessageFilter {
	return func(u *adapter.Update) bool {
		return u.DeletedChannelMessages != nil && u.DeletedChannelMessages.ChannelID == channelID
	}
}
