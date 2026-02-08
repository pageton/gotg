package adapter

import (
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

func (u *Update) HasMessage() bool {
	return u.EffectiveMessage != nil && u.EffectiveMessage.ID != 0
}

// IsBot returns true if the current session belongs to a bot account.
func (u *Update) IsBot() bool {
	return u.Self != nil && u.Self.Bot
}

// Args parses and returns the arguments from the update.
// For messages, splits the message text by whitespace.
// For callback queries, splits the callback data by whitespace.
// For inline queries, splits the query text by whitespace.
// Returns an empty slice if no applicable content exists.
func (u *Update) Args() []string {
	switch {
	case u.HasMessage():
		return strings.Fields(u.EffectiveMessage.Text)
	case u.CallbackQuery != nil:
		return strings.Fields(string(u.CallbackQuery.Data))
	case u.InlineQuery != nil:
		return strings.Fields(u.InlineQuery.Query())
	default:
		return make([]string, 0)
	}
}

// EffectiveInlineQuery returns the wrapped InlineQuery for the current update.
// Returns nil if the update is not an inline query.
func (u *Update) EffectiveInlineQuery() *types.InlineQuery {
	return u.InlineQuery
}

// EffectiveChosenInlineResult returns the wrapped ChosenInlineResult for the current update.
// Returns nil if the update is not a chosen inline result.
func (u *Update) EffectiveChosenInlineResult() *types.ChosenInlineResult {
	return u.ChosenInlineResult
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
		// Fallback 1: try to construct minimal user from PeerStorage
		if u.Ctx.PeerStorage != nil {
			storagePeer := u.Ctx.PeerStorage.GetPeerByID(u.userID)
			if storagePeer != nil && storagePeer.Type == 1 { // 1 = TypeUser
				// Construct minimal user with ID and AccessHash from storage
				tgUser = &tg.User{
					ID:         u.userID,
					AccessHash: storagePeer.AccessHash,
				}
			}
		}
		// Fallback 2: create stub user with just ID (for outgoing messages where recipient isn't in entities)
		if tgUser == nil {
			tgUser = &tg.User{
				ID: u.userID,
			}
		}
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
	var peer tg.PeerClass
	switch {
	case u.HasMessage():
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
		// Fallback: create stub chat with just ID (for chats not in entities)
		tgChat = &tg.Chat{
			ID: c.ChatID,
		}
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
	var peer tg.PeerClass
	switch {
	case u.HasMessage():
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
		// Fallback 1: try to construct minimal channel from PeerStorage
		if u.Ctx.PeerStorage != nil {
			// Convert plain channel ID to TDLib format for lookup
			// PeerStorage stores channels with -100 prefix (TDLib format)
			// Create a temporary types.Channel to get the properly encoded ID
			tempChannel := types.Channel{
				Channel:     tg.Channel{ID: c.ChannelID},
				PeerStorage: u.Ctx.PeerStorage,
			}
			encodedID := tempChannel.GetID()
			storagePeer := u.Ctx.PeerStorage.GetPeerByID(encodedID)
			if storagePeer != nil && storagePeer.Type == 3 { // 3 = TypeChannel
				// Construct minimal channel with ID and AccessHash from storage
				tgChannel = &tg.Channel{
					ID:         c.ChannelID,
					AccessHash: storagePeer.AccessHash,
				}
			}
		}
		// Fallback 2: create stub channel with just ID (for channels not in entities/storage)
		if tgChannel == nil {
			tgChannel = &tg.Channel{
				ID: c.ChannelID,
			}
		}
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
	var peer tg.PeerClass
	switch {
	case u.HasMessage():
		peer = u.EffectiveMessage.PeerID
	case u.CallbackQuery != nil:
		peer = u.CallbackQuery.Peer
	case u.ChatJoinRequest != nil:
		peer = u.ChatJoinRequest.Peer
	case u.ChatParticipant != nil:
		peer = &tg.PeerUser{UserID: u.ChatParticipant.UserID}
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
		// Fallback 1: try to construct minimal user from PeerStorage
		if u.Ctx.PeerStorage != nil {
			storagePeer := u.Ctx.PeerStorage.GetPeerByID(c.UserID)
			if storagePeer != nil && storagePeer.Type == 1 { // 1 = TypeUser
				// Construct minimal user with ID and AccessHash from storage
				tgUser = &tg.User{
					ID:         c.UserID,
					AccessHash: storagePeer.AccessHash,
				}
			}
		}
		// Fallback 2: create stub user with just ID (for users not in entities/storage)
		if tgUser == nil {
			tgUser = &tg.User{
				ID: c.UserID,
			}
		}
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
	if !u.HasMessage() {
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
	if !u.HasMessage() {
		return false
	}
	return u.EffectiveMessage.ReplyTo != nil
}

// IsOutgoing returns true if the effective message was sent by this client (self).
// Works for both user accounts (via tg.Message.Out) and bot accounts
// (via FromID == Self.ID fallback, since bots always receive Out=false).
func (u *Update) IsOutgoing() bool {
	if !u.HasMessage() {
		return false
	}
	return u.EffectiveMessage.IsOutgoing()
}

// IsIncoming returns true if the effective message was sent by another user
// (not by this client). This is the inverse of IsOutgoing.
func (u *Update) IsIncoming() bool {
	return !u.IsOutgoing()
}

// ConnectionID returns the business connection ID for this update.
// Returns empty string if the update is not business-related.
func (u *Update) ConnectionID() string {
	switch {
	case u.BusinessConnection != nil:
		return u.BusinessConnection.Connection.ConnectionID
	case u.BusinessMessage != nil:
		return u.BusinessMessage.ConnectionID
	case u.BusinessEditedMessage != nil:
		return u.BusinessEditedMessage.ConnectionID
	case u.BusinessDeletedMessages != nil:
		return u.BusinessDeletedMessages.ConnectionID
	case u.BusinessCallbackQuery != nil:
		return u.BusinessCallbackQuery.ConnectionID
	default:
		return ""
	}
}

// IsBusinessUpdate returns true if this update originated from a business connection.
func (u *Update) IsBusinessUpdate() bool {
	return u.ConnectionID() != ""
}

// fillUserIDFromMessage populates the userID field by extracting the user ID
// from various update types. Used internally during update construction.
func (u *Update) fillUserIDFromMessage(selfUserID int64) {
	if u.HasMessage() {
		m := u.EffectiveMessage
		// For incoming messages, get user from FromID
		if userPeer, ok := m.FromID.(*tg.PeerUser); ok {
			u.userID = userPeer.UserID
			return
		}
		// For outgoing messages or private messages, get user from PeerID
		if userPeer, ok := m.PeerID.(*tg.PeerUser); ok {
			u.userID = userPeer.UserID
			return
		}
	}
	// Fallback: try to find a non-self user in entities
	// Note: This should only be used if the above methods fail
	if u.Entities != nil && u.Entities.Users != nil {
		for uID := range u.Entities.Users {
			if uID == selfUserID {
				continue
			}
			u.userID = uID
			return
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
	case u.HasMessage():
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
	case u.HasMessage():
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
	if !u.HasMessage() {
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
