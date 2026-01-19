package filters

import "github.com/pageton/gotg/adapter"

type chatMemberUpdated struct{}

// All returns true on every type of tg.UpdateChannelParticipant and tg.UpdateChatParticipant update.
func (*chatMemberUpdated) All(_ *adapter.Update) bool {
	return true
}

// ChatUpdate returns true on every type of tg.UpdateChatParticipant update.
func (*chatMemberUpdated) ChatUpdate(u *adapter.Update) bool {
	return u.ChatParticipant != nil
}

// ChannelUpdate returns true on every type of tg.UpdateChannelParticipant update.
func (*chatMemberUpdated) ChannelUpdate(u *adapter.Update) bool {
	return u.ChannelParticipant != nil
}

// FromUserID checks if the tg.UpdateChatParticipant and tg.UpdateChannelParticipant was sent by the provided user id and returns true if matches.
func (*chatMemberUpdated) FromUserID(userID int64) ChatMemberUpdatedFilter {
	return func(u *adapter.Update) bool {
		if u.ChannelParticipant != nil {
			return u.ChannelParticipant.UserID == userID
		}
		if u.ChatParticipant != nil {
			return u.ChatParticipant.UserID == userID
		}
		return false
	}
}

// FromChatID checks if the tg.UpdateChatParticipant and tg.UpdateChannelParticipant was sent at the provided chat id and returns true if matches.
func (*chatMemberUpdated) FromChatID(chatID int64) ChatMemberUpdatedFilter {
	return func(u *adapter.Update) bool {
		if u.ChannelParticipant != nil {
			return u.ChannelParticipant.ChannelID == chatID
		}
		if u.ChatParticipant != nil {
			return u.ChatParticipant.ChatID == chatID
		}
		return false
	}
}
