package adapter

import (
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

// Args parses and returns the arguments from the update.
// For messages, splits the message text by whitespace.
// For callback queries, splits the callback data by whitespace.
// For inline queries, splits the query text by whitespace.
// Returns an empty slice if no applicable content exists.
func (u *Update) Args() []string {
	switch {
	case u.EffectiveMessage != nil:
		return strings.Fields(u.EffectiveMessage.Text)
	case u.CallbackQuery != nil:
		return strings.Fields(string(u.CallbackQuery.Data))
	case u.InlineQuery != nil:
		return strings.Fields(u.InlineQuery.Query)
	default:
		return make([]string, 0)
	}
}

// EffectiveUser returns the types.User who is responsible for the update.
func (u *Update) EffectiveUser() *types.User {
	if u.Entities == nil {
		return nil
	}
	if u.userID == 0 {
		return nil
	}
	tgUser := u.Entities.Users[u.userID]
	if tgUser == nil {
		return nil
	}
	return &types.User{
		User:        *tgUser,
		Ctx:         u.Ctx.Context,
		RawClient:   u.Ctx.Raw,
		PeerStorage: u.Ctx.PeerStorage,
		SelfID:      u.Ctx.Self.ID,
	}
}

// GetChat returns the responsible types.Chat for the current update.
func (u *Update) GetChat() *types.Chat {
	if u.Entities == nil {
		return nil
	}
	var (
		peer tg.PeerClass
	)
	switch {
	case u.EffectiveMessage != nil:
		peer = u.EffectiveMessage.PeerID
	case u.CallbackQuery != nil:
		peer = u.CallbackQuery.Peer
	case u.ChatJoinRequest != nil:
		peer = u.ChatJoinRequest.Peer
	case u.ChatParticipant != nil:
		peer = &tg.PeerChat{ChatID: u.ChatParticipant.ChatID}
	}
	if peer == nil {
		return nil
	}
	c, ok := peer.(*tg.PeerChat)
	if !ok {
		return nil
	}
	tgChat := u.Entities.Chats[c.ChatID]
	if tgChat == nil {
		return nil
	}
	chat := types.Chat{
		Chat:        *tgChat,
		Ctx:         u.Ctx.Context,
		RawClient:   u.Ctx.Raw,
		PeerStorage: u.Ctx.PeerStorage,
		SelfID:      u.Ctx.Self.ID,
	}
	return &chat
}

// GetChannel returns the responsible types.Channel for the current update.
func (u *Update) GetChannel() *types.Channel {
	if u.Entities == nil {
		return nil
	}
	var (
		peer tg.PeerClass
	)
	switch {
	case u.EffectiveMessage != nil:
		peer = u.EffectiveMessage.PeerID
	case u.CallbackQuery != nil:
		peer = u.CallbackQuery.Peer
	case u.ChatJoinRequest != nil:
		peer = u.ChatJoinRequest.Peer
	case u.ChannelParticipant != nil:
		peer = &tg.PeerChannel{ChannelID: u.ChannelParticipant.ChannelID}
	}
	if peer == nil {
		return nil
	}
	c, ok := peer.(*tg.PeerChannel)
	if !ok {
		return nil
	}
	tgChannel := u.Entities.Channels[c.ChannelID]
	if tgChannel == nil {
		return nil
	}
	channel := types.Channel{
		Channel:     *tgChannel,
		Ctx:         u.Ctx.Context,
		RawClient:   u.Ctx.Raw,
		PeerStorage: u.Ctx.PeerStorage,
		SelfID:      u.Ctx.Self.ID,
	}
	return &channel
}

// GetUserChat returns the responsible types.User for the current update.
func (u *Update) GetUserChat() *types.User {
	if u.Entities == nil {
		return nil
	}
	var (
		peer tg.PeerClass
	)
	switch {
	case u.EffectiveMessage != nil:
		peer = u.EffectiveMessage.PeerID
	case u.CallbackQuery != nil:
		peer = u.CallbackQuery.Peer
	case u.ChatJoinRequest != nil:
		peer = u.ChatJoinRequest.Peer
	case u.ChatParticipant != nil:
		peer = &tg.PeerChat{ChatID: u.ChatParticipant.ChatID}
	}
	if peer == nil {
		return nil
	}
	c, ok := peer.(*tg.PeerUser)
	if !ok {
		return nil
	}
	tgUser := u.Entities.Users[c.UserID]
	if tgUser == nil {
		return nil
	}
	return &types.User{
		User:        *tgUser,
		Ctx:         u.Ctx.Context,
		RawClient:   u.Ctx.Raw,
		PeerStorage: u.Ctx.PeerStorage,
		SelfID:      u.Ctx.Self.ID,
	}
}

// EffectiveChat returns the responsible EffectiveChat for the current update.
func (u *Update) EffectiveChat() types.EffectiveChat {
	if c := u.GetChannel(); c != nil {
		return c
	}
	if c := u.GetChat(); c != nil {
		return c
	}
	if c := u.GetUserChat(); c != nil {
		return c
	}
	return &types.EmptyUC{}
}

// EffectiveReply returns the message that this message is replying to.
// It lazily fetches the reply message if not already populated.
func (u *Update) EffectiveReply() *types.Message {
	if u.EffectiveMessage == nil {
		return nil
	}

	if u.EffectiveMessage.ReplyToMessage != nil {
		return u.EffectiveMessage.ReplyToMessage
	}

	_ = u.EffectiveMessage.SetRepliedToMessage(u.Ctx.Context, u.Ctx.Raw, u.Ctx.PeerStorage)
	return u.EffectiveMessage.ReplyToMessage
}

// IsReply returns true if the effective message is a reply to another message.
func (u *Update) IsReply() bool {
	if u.EffectiveMessage == nil {
		return false
	}
	return u.EffectiveMessage.ReplyTo != nil
}

// fillUserIDFromMessage populates the userID field by extracting the user ID
// from various update types. Used internally during update construction.
func (u *Update) fillUserIDFromMessage(selfUserID int64) {
	if m := u.EffectiveMessage; m != nil {
		if userPeer, ok := m.FromID.(*tg.PeerUser); ok {
			u.userID = userPeer.UserID
			return
		}
		if userPeer, ok := m.PeerID.(*tg.PeerUser); ok {
			u.userID = userPeer.UserID
			return
		}
	}
	if u.Entities != nil && u.Entities.Users != nil {
		for uID := range u.Entities.Users {
			if uID == selfUserID {
				continue
			}
			u.userID = uID
			break
		}
	}
}

// ChatID returns the chat ID for this update.
// For messages and callback queries, extracts from the peer ID.
// Returns 0 if no chat can be determined.
func (u *Update) ChatID() int64 {
	if chat := u.EffectiveChat(); chat != nil {
		return chat.GetID()
	}
	if u.CallbackQuery != nil && u.CallbackQuery.Peer != nil {
		switch peer := u.CallbackQuery.Peer.(type) {
		case *tg.PeerUser:
			return peer.UserID
		case *tg.PeerChat:
			return peer.ChatID
		case *tg.PeerChannel:
			return peer.ChannelID
		}
	}
	return 0
}

// ChannelID returns the channel ID for this update.
// For messages and callback queries, extracts from the peer ID.
// Returns 0 if no channel can be determined.
func (u *Update) ChannelID() int64 {
	if channel := u.GetChannel(); channel != nil {
		return channel.GetID()
	}
	if u.CallbackQuery != nil && u.CallbackQuery.Peer != nil {
		if peer, ok := u.CallbackQuery.Peer.(*tg.PeerChannel); ok {
			return peer.ChannelID
		}
	}
	if u.ChannelParticipant != nil {
		return u.ChannelParticipant.ChannelID
	}
	return 0
}

// MsgID returns the message ID for this update.
// For messages, returns the message ID.
// For callback queries, returns the message ID that triggered the callback.
// Returns 0 if no message ID exists.
func (u *Update) MsgID() int {
	switch {
	case u.EffectiveMessage != nil:
		return u.EffectiveMessage.ID
	case u.CallbackQuery != nil:
		return u.CallbackQuery.MsgID
	default:
		return 0
	}
}

// MessageID returns the message ID for this update (alias for MsgID).
func (u *Update) MessageID() int {
	switch {
	case u.EffectiveMessage != nil:
		return u.EffectiveMessage.ID
	case u.CallbackQuery != nil:
		return u.CallbackQuery.MsgID
	default:
		return 0
	}
}

// UserID returns the effective user ID for this update.
// Returns 0 if no user can be determined.
func (u *Update) UserID() int64 {
	if user := u.EffectiveUser(); user != nil {
		return user.GetID()
	}
	return 0
}

// FirstName returns the first name of the effective user.
// Returns an empty string if no user exists.
func (u *Update) FirstName() string {
	if user := u.EffectiveUser(); user != nil {
		return user.FirstName
	}
	return ""
}

// LastName returns the last name of the effective user.
// Returns an empty string if no user exists.
func (u *Update) LastName() string {
	if user := u.EffectiveUser(); user != nil {
		return user.LastName
	}
	return ""
}

// FullName returns the full name (first name + last name) of the effective user.
// Returns an empty string if no user exists.
func (u *Update) FullName() string {
	if user := u.EffectiveUser(); user != nil {
		return user.FirstName + " " + user.LastName
	}
	return ""
}

// Username returns the username of the effective user.
// Returns an empty string if no user exists or username is not set.
func (u *Update) Username() string {
	if user := u.EffectiveUser(); user != nil {
		return user.Username
	}
	return ""
}

// Usernames returns all usernames (including collected) of the effective user.
// Returns nil if no user exists.
func (u *Update) Usernames() []tg.Username {
	if user := u.EffectiveUser(); user != nil {
		return user.Usernames
	}
	return nil
}

// Text returns the text content of the effective message.
// Returns an empty string if no message exists.
func (u *Update) Text() string {
	if u.EffectiveMessage == nil {
		return ""
	}
	return u.EffectiveMessage.Text
}

// LangCode returns the language code of the effective user.
// Returns an empty string if no user exists.
func (u *Update) LangCode() string {
	if user := u.EffectiveUser(); user != nil {
		return user.LangCode
	}
	return ""
}
