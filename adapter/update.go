package adapter

import (
	"fmt"
	"html"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/parsemode"
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

	// If ReplyToMessage is already populated, return it
	if u.EffectiveMessage.ReplyToMessage != nil {
		return u.EffectiveMessage.ReplyToMessage
	}

	// Otherwise, fetch and populate the reply message
	_ = u.EffectiveMessage.SetRepliedToMessage(u.Ctx.Context, u.Ctx.Raw, u.Ctx.PeerStorage)
	return u.EffectiveMessage.ReplyToMessage
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
			name = html.EscapeString(v)
		}
	} else if len(args) >= 2 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		}
		if n, ok := args[1].(string); ok {
			name = html.EscapeString(n)
		}
	}

	return fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", userID, name)
}

// Delete deletes the effective message for this update.
// Returns an error if the deletion fails.
func (u *Update) Delete() error {
	return u.Ctx.DeleteMessages(u.ChatID(), []int{u.MsgID()})
}

// GetUser fetches full user information for the effective user.
// Returns nil if no user exists or on error.
func (u *Update) GetUser() (*tg.UserFull, error) {
	return u.Ctx.GetUser(u.UserID())
}

// Pin pins the effective message in the chat.
// Returns updates confirming the action or an error.
func (u *Update) Pin() (tg.UpdatesClass, error) {
	return u.Ctx.PinMessage(u.ChatID(), u.MsgID())
}

// Unpin unpins the effective message in the chat.
// Returns an error if the operation fails.
func (u *Update) Unpin() error {
	return u.Ctx.UnpinMessage(u.ChatID(), u.MsgID())
}

// UnpinAll unpins all messages in the current chat.
// Returns an error if the operation fails.
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

// SetLang sets the language preference for the effective user.
// This requires i18n middleware to be initialized.
//
// Parameters:
//   - lang: The language code (e.g., "en", "es") or language.Tag
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

// SendMessage sends a text message to the specified chat.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
//
// NOTE: This method does NOT reply by default. To reply to a specific message,
// set ReplyMessageID in ReplyOpts.
//
// Parameters:
//   - chatID: The target chat ID (use 0 to use the current update's chat)
//   - text: The message text
//   - opts: Optional ReplyOpts for formatting, reply markup, etc.
//
// Returns the sent Message or an error.
//
// Example:
//
//	msg, err := u.SendMessage(0, "Hello, world!")  // Send to current chat (no reply)
//	msg, err := u.SendMessage(chatID, "<b>Bold text</b>", &ReplyOpts{
//	    ParseMode: "HTML",
//	})
//	// Reply to a specific message:
//	msg, err := u.SendMessage(chatID, "Reply text", &ReplyOpts{
//	    ReplyMessageID: 123,
//	})
func (u *Update) SendMessage(chatID int64, text string, opts ...*ReplyOpts) (*types.Message, error) {
	if chatID == 0 {
		chatID = u.ChatID()
	}
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	// Build request from opts
	var opt *ReplyOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &ReplyOpts{}
	}

	// Default parse mode is HTML
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Parse HTML/Markdown to entities
	var messageText string
	var entities []tg.MessageEntityClass

	if parseMode != ModeNone {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case HTML:
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(text, mode)
		if err == nil && result != nil {
			messageText = result.Text
			entities = result.Entities
		} else {
			messageText = text
		}
	} else {
		messageText = text
	}

	// Build the send request
	req := &tg.MessagesSendMessageRequest{
		Message:  messageText,
		Entities: entities,
	}

	// Set reply markup
	if opt.Markup != nil {
		req.ReplyMarkup = opt.Markup
	}

	// Set flags
	if opt.NoWebpage {
		req.Flags |= 4
		req.NoWebpage = true
	}
	if opt.Silent {
		req.Silent = true
	}
	if opt.ClearDraft {
		req.ClearDraft = true
	}
	if opt.NoForwards {
		req.Noforwards = true
	}
	if opt.ScheduleDate != 0 {
		req.ScheduleDate = int(opt.ScheduleDate)
	}
	if opt.Effect != 0 {
		req.Effect = opt.Effect
	}

	// Add ReplyTo if ReplyMessageID is set
	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	}

	return u.Ctx.SendMessage(chatID, req)
}

// SendMedia sends media (photo, document, video, etc.) to the chat associated with this update.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
// Default parse mode for caption is HTML.
//
// NOTE: This method does NOT reply by default. To reply to a specific message,
// set ReplyMessageID in ReplyMediaOpts.
//
// Parameters:
//   - media: The media to send (tg.InputMediaClass)
//   - caption: Optional caption text
//   - opts: Optional ReplyMediaOpts
//
// Returns the sent Message or an error.
//
// Example using InputMedia:
//
//	msg, err := u.SendMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, "Photo caption")
//
// Example using fileID (convert with types.InputMediaFromFileID):
//
//	media, _ := types.InputMediaFromFileID(fileID, "caption")
//	msg, err := u.SendMedia(media, "caption")
//
// Example replying to a specific message:
//
//	msg, err := u.SendMedia(media, "caption", &ReplyMediaOpts{
//	    ReplyMessageID: 123,
//	})
func (u *Update) SendMedia(media tg.InputMediaClass, caption string, opts ...*ReplyMediaOpts) (*types.Message, error) {
	chatID := u.ChatID()
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	var opt *ReplyMediaOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &ReplyMediaOpts{}
	}

	// Default parse mode is HTML
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// If caption passed directly, use it
	if caption != "" && opt.Caption == "" {
		opt.Caption = caption
	}

	// Parse caption for entities
	var captionText string
	var entities []tg.MessageEntityClass

	if opt.Caption != "" && parseMode != ModeNone {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case HTML:
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(opt.Caption, mode)
		if err == nil && result != nil {
			captionText = result.Text
			entities = result.Entities
		} else {
			captionText = opt.Caption
		}
	} else {
		captionText = opt.Caption
	}

	// Build the send media request
	req := &tg.MessagesSendMediaRequest{
		Media:    media,
		Message:  captionText,
		Entities: entities,
	}

	// Set reply markup
	if opt.Markup != nil {
		req.ReplyMarkup = opt.Markup
	}

	// Set flags
	if opt.Silent {
		req.Silent = true
	}
	if opt.ClearDraft {
		req.ClearDraft = true
	}
	if opt.NoForwards {
		req.Noforwards = true
	}
	if opt.ScheduleDate != 0 {
		req.ScheduleDate = int(opt.ScheduleDate)
	}
	if opt.InvertMedia {
		req.InvertMedia = true
	}

	// Add ReplyTo if ReplyMessageID is set
	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	}

	return u.Ctx.SendMedia(chatID, req)
}

// SendMultiMedia sends multiple media items as an album to the chat associated with this update.
// Albums can contain up to 10 media items.
//
// NOTE: This method does NOT reply by default. To reply to a specific message,
// set ReplyMessageID in ReplyMediaOpts.
//
// Parameters:
//   - media: Slice of InputMediaClass items
//   - opts: Optional ReplyMediaOpts (applied to all items)
//
// Returns the sent Message or an error.
//
// Example:
//
//	msgs, err := u.SendMultiMedia([]tg.InputMediaClass{
//	    &tg.InputMediaPhoto{ID: &tg.InputPhoto{...}},
//	    &tg.InputMediaPhoto{ID: &tg.InputPhoto{...}},
//	}, nil)
//
// Example replying to a specific message:
//
//	msgs, err := u.SendMultiMedia(media, &ReplyMediaOpts{
//	    ReplyMessageID: 123,
//	})
func (u *Update) SendMultiMedia(media []tg.InputMediaClass, opts ...*ReplyMediaOpts) (*types.Message, error) {
	chatID := u.ChatID()
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	if len(media) == 0 {
		return nil, fmt.Errorf("media slice cannot be empty")
	}

	var opt *ReplyMediaOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &ReplyMediaOpts{}
	}

	// Default parse mode is HTML
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Build single media request with array
	// Note: SendMultiMedia uses MessagesSendMultiMediaRequest

	inputMedia := make([]tg.InputSingleMedia, len(media))
	for i, m := range media {
		var caption string
		var entities []tg.MessageEntityClass

		// Parse caption if provided
		if opt.Caption != "" && parseMode != ModeNone {
			var mode parsemode.ParseMode
			switch strings.ToUpper(strings.TrimSpace(parseMode)) {
			case HTML:
				mode = parsemode.ModeHTML
			case "MARKDOWN", "MARKDOWNV2":
				mode = parsemode.ModeMarkdown
			default:
				mode = parsemode.ModeNone
			}

			result, err := parsemode.Parse(opt.Caption, mode)
			if err == nil && result != nil {
				caption = result.Text
				entities = result.Entities
			} else {
				caption = opt.Caption
			}
		}

		inputMedia[i] = tg.InputSingleMedia{
			Media:    m,
			RandomID: u.Ctx.generateRandomID(),
			Message:  caption,
			Entities: entities,
		}
	}

	req := &tg.MessagesSendMultiMediaRequest{
		MultiMedia: inputMedia,
	}

	// Set flags
	if opt.Silent {
		req.Silent = true
	}
	if opt.ClearDraft {
		req.ClearDraft = true
	}
	if opt.NoForwards {
		req.Noforwards = true
	}
	if opt.ScheduleDate != 0 {
		req.ScheduleDate = int(opt.ScheduleDate)
	}

	// Add ReplyTo if ReplyMessageID is set
	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	}

	return u.Ctx.SendMultiMedia(chatID, req)
}

// EditMessage edits a message in the specified chat.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
//
// Parameters:
//   - chatID: The target chat ID (use 0 to use the current update's chat)
//   - messageID: The ID of the message to edit
//   - text: New message text
//   - opts: Optional ReplyOpts for entities, reply markup, etc.
//
// Returns the edited Message or an error.
//
// Example:
//
//	msg, err := u.EditMessage(0, 123, "Updated text")  // Edit in current chat
//	msg, err := u.EditMessage(chatID, 123, "<b>Updated</b>", &ReplyOpts{
//	    ParseMode: "HTML",
//	})
func (u *Update) EditMessage(chatID int64, messageID int, text string, opts ...*ReplyOpts) (*types.Message, error) {
	if chatID == 0 {
		chatID = u.ChatID()
	}
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	var opt *ReplyOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &ReplyOpts{}
	}

	// Default parse mode is HTML
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Parse HTML/Markdown to entities
	var messageText string
	var entities []tg.MessageEntityClass

	if parseMode != ModeNone {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case HTML:
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(text, mode)
		if err == nil && result != nil {
			messageText = result.Text
			entities = result.Entities
		} else {
			messageText = text
		}
	} else {
		messageText = text
	}

	// Build the edit request
	req := &tg.MessagesEditMessageRequest{
		ID:        messageID,
		Message:   messageText,
		Entities:  entities,
		NoWebpage: opt.NoWebpage,
	}

	// Set reply markup
	if opt.Markup != nil {
		req.ReplyMarkup = opt.Markup
	}

	return u.Ctx.EditMessage(chatID, req)
}

// EditMessageMedia edits the media of a specific message in the specified chat.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
// This differs from EditMedia() which edits the current update's effective message.
//
// Parameters:
//   - chatID: The target chat ID (use 0 to use the current update's chat)
//   - messageID: The ID of the message to edit
//   - media: The new media (tg.InputMediaClass)
//   - caption: New caption text
//   - opts: Optional ReplyMediaOpts
//
// Returns the edited Message or an error.
//
// Example using InputMedia:
//
//	msg, err := u.EditMessageMedia(0, 123, &tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, "New caption")  // Edit in current chat
//
// Example using fileID (convert with types.InputMediaFromFileID):
//
//	media, _ := types.InputMediaFromFileID(fileID, "caption")
//	msg, err := u.EditMessageMedia(chatID, 123, media, "caption")
func (u *Update) EditMessageMedia(chatID int64, messageID int, media tg.InputMediaClass, caption string, opts ...*ReplyMediaOpts) (*types.Message, error) {
	if chatID == 0 {
		chatID = u.ChatID()
	}
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	var opt *ReplyMediaOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = &ReplyMediaOpts{}
	}

	// Default parse mode is HTML
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// If caption passed directly, use it
	if caption != "" && opt.Caption == "" {
		opt.Caption = caption
	}

	// Parse caption for entities
	var captionText string
	var entities []tg.MessageEntityClass

	if opt.Caption != "" && parseMode != ModeNone {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case HTML:
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(opt.Caption, mode)
		if err == nil && result != nil {
			captionText = result.Text
			entities = result.Entities
		} else {
			captionText = opt.Caption
		}
	} else {
		captionText = opt.Caption
	}

	// Build the edit request
	req := &tg.MessagesEditMessageRequest{
		ID:       messageID,
		Media:    media,
		Message:  captionText,
		Entities: entities,
	}

	// Set reply markup
	if opt.Markup != nil {
		req.ReplyMarkup = opt.Markup
	}

	return u.Ctx.EditMessage(chatID, req)
}
