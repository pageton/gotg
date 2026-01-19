package adapter

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

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

	// If ReplyToMessage is already populated, return it
	if u.EffectiveMessage.ReplyToMessage != nil {
		return u.EffectiveMessage.ReplyToMessage
	}

	// Otherwise, fetch and populate the reply message
	_ = u.EffectiveMessage.SetRepliedToMessage(u.Ctx.Context, u.Ctx.Raw, u.Ctx.PeerStorage)
	return u.EffectiveMessage.ReplyToMessage
}

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

func (u *Update) ChatID() int64 {
	if chat := u.EffectiveChat(); chat != nil {
		return chat.GetID()
	}
	// Fallback for callback queries - extract ID directly from peer
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

func (u *Update) UserID() int64 {
	if user := u.EffectiveUser(); user != nil {
		return user.GetID()
	}
	return 0
}

func (u *Update) FirstName() string {
	if user := u.EffectiveUser(); user != nil {
		return user.FirstName
	}
	return ""
}

func (u *Update) LastName() string {
	if user := u.EffectiveUser(); user != nil {
		return user.LastName
	}
	return ""
}

func (u *Update) FullName() string {
	if user := u.EffectiveUser(); user != nil {
		return user.FirstName + " " + user.LastName
	}
	return ""
}

func (u *Update) Username() string {
	if user := u.EffectiveUser(); user != nil {
		return user.Username
	}
	return ""
}

func (u *Update) Usernames() []tg.Username {
	if user := u.EffectiveUser(); user != nil {
		return user.Usernames
	}
	return nil
}

func (u *Update) Text() string {
	return u.EffectiveMessage.Text
}

func (u *Update) LangCode() string {
	if user := u.EffectiveUser(); user != nil {
		return user.LangCode
	}
	return ""
}

// Mention generates an HTML mention link for a Telegram user.
//
// Behavior:
// - No arguments: uses the Update's default UserID() and FullName().
// - One argument:
//   - int/int64 → overrides userID, keeps default name.
//   - string → overrides name, keeps default userID.
//
// - Two arguments: first is userID (int/int64), second is name (string).
// - The name can be any string, including numeric names.
// - Returns a string in the format: <a href='tg://user?id=USERID'>NAME</a>
func (u *Update) Mention(args ...any) string {
	userID := u.UserID()
	name := u.FullName()

	if len(args) == 1 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		case string:
			name = v
		}
	} else if len(args) >= 2 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		}
		if n, ok := args[1].(string); ok {
			name = n
		}
	}

	return fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", userID, name)
}

func (u *Update) Delete() error {
	return u.Ctx.DeleteMessages(u.ChatID(), []int{u.MsgID()})
}

func (u *Update) GetUser() (*tg.UserFull, error) {
	return u.Ctx.GetUser(u.UserID())
}

func (u *Update) Pin() (tg.UpdatesClass, error) {
	return u.Ctx.PinMessage(u.ChatID(), u.MsgID())
}

// Unpin unpins the effective message in the chat.
func (u *Update) Unpin() error {
	return u.Ctx.UnpinMessage(u.ChatID(), u.MsgID())
}

// UnpinAll unpins all messages in the current chat.
func (u *Update) UnpinAll() error {
	return u.Ctx.UnpinAllMessages(u.ChatID())
}

// Answer answers the callback query.
// text: The notification text (use empty string for silent).
// opts: Optional *CallbackOptions for alert, cacheTime, url.
//
// Example:
//
//	u.Answer("Done!", nil)
//	u.Answer("Error!", &CallbackOptions{Alert: true})
func (u *Update) Answer(text string, opts ...*CallbackOptions) (bool, error) {
	if u.CallbackQuery == nil {
		return false, fmt.Errorf("no callback query in this update")
	}

	// Default values
	alert := false
	cacheTime := 0
	url := ""

	if len(opts) > 0 && opts[0] != nil {
		alert = opts[0].Alert
		cacheTime = opts[0].CacheTime
		url = opts[0].URL
	}

	return u.Ctx.Raw.MessagesSetBotCallbackAnswer(u.Ctx, &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID:   u.CallbackQuery.QueryID,
		Message:   text,
		Alert:     alert,
		CacheTime: cacheTime,
		URL:       url,
	})
}

// T returns a translation for the given key.
// Supports both simple args and context (Args) for pluralization, gender, etc.
// This method requires i18n middleware to be initialized.
//
// Examples:
//
//	// Simple translation with positional args
//	text := u.T("greeting", userName)
//
//	// Translation with context (pluralization, gender)
//	text := u.T("items_count", &i18n.Args{Count: 5})
//
//	// Translation with named args
//	text := u.T("welcome", &i18n.Args{Args: map[string]any{"name": userName}})
func (u *Update) T(key string, args ...any) string {
	if u.Ctx == nil {
		return key
	}
	// Get translator from middleware
	// The middleware stores a reference that we can access
	return updateTImpl(u, key, args...)
}

func (u *Update) SetLang(lang any) {
	updateSetLangImpl(u, lang)
}

// GetLang returns the user's current language preference.
// This method requires i18n middleware to be initialized.
//
// Example:
//
//	lang := u.GetLang()
func (u *Update) GetLang() any {
	return updateGetLangImpl(u)
}
