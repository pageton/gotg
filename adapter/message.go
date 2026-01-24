package adapter

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	mtp_errors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/types"
)

// ReplyOpts contains optional parameters for Reply method.
type ReplyOpts struct {
	// ParseMode is the parse mode for formatting (default: HTML).
	// Use ext.HTML, ext.Markdown, ext.MarkdownV2, or ext.ModeNone.
	ParseMode string
	// NoWebpage disables link preview
	NoWebpage bool
	// Silent sends the message silently
	Silent bool
	// Markup is the inline/reply keyboard
	Markup tg.ReplyMarkupClass
	// ClearDraft clears saved draft in this chat after sending
	ClearDraft bool
	// LinkPreview generates link preview for URLs in message
	LinkPreview bool
	// NoForwards restricts forwarding and saving of this message
	NoForwards bool
	// ReplyTo is the full reply configuration with quote support
	ReplyTo *tg.InputReplyToMessage
	// TopicID is the forum topic ID to send message in
	TopicID int64
	// ScheduleDate is Unix timestamp to schedule message delivery
	ScheduleDate int32
	// SendAs is send as channel or linked chat (peer ID)
	SendAs int64
	// TTL is self-destruct timer in seconds (for secret media)
	TTL int32
	// Effect is message effect animation ID
	Effect int64
	// WithoutReply sends message without replying to the original message
	WithoutReply bool
	// ReplyMessageID is the message ID to reply to (instead of the current message)
	ReplyMessageID int
}

// ReplyMediaOpts contains optional parameters for ReplyMedia method.
type ReplyMediaOpts struct {
	// Caption is the caption for the media
	Caption string
	// ParseMode is the parse mode for caption formatting (default: HTML).
	// Use ext.HTML, ext.Markdown, ext.MarkdownV2, or ext.ModeNone.
	ParseMode string
	// Markup is the inline/reply keyboard
	Markup tg.ReplyMarkupClass
	// Silent sends the message silently
	Silent bool
	// Attributes are document attributes like filename, dimensions, duration
	Attributes []tg.DocumentAttributeClass
	// MimeType is MIME type override (e.g., "video/mp4", "audio/mpeg")
	MimeType string
	// ClearDraft clears saved draft in this chat after sending
	ClearDraft bool
	// Entities are pre-parsed formatting entities (overrides ParseMode)
	Entities []tg.MessageEntityClass
	// FileName is custom filename for the uploaded file
	FileName string
	// ForceDocument sends media as document instead of embedded preview
	ForceDocument bool
	// InvertMedia displays media below caption instead of above
	InvertMedia bool
	// LinkPreview generates link preview for URLs in caption
	LinkPreview bool
	// NoForwards restricts forwarding and saving of this message
	NoForwards bool
	// NoSoundVideo strips audio track from video file
	NoSoundVideo bool
	// ReplyTo is the full reply configuration with quote support
	ReplyTo *tg.InputReplyToMessage
	// TopicID is the forum topic ID to send message in
	TopicID int64
	// UpdateStickerOrder moves used sticker to top of recent stickers
	UpdateStickerOrder bool
	// ScheduleDate is Unix timestamp to schedule message delivery
	ScheduleDate int32
	// ScheduleRepeatPeriod is repeat interval in seconds for scheduled message
	ScheduleRepeatPeriod int32
	// SendAs is send as channel or linked chat (peer ID)
	SendAs int64
	// Thumb is custom thumbnail (file path or bytes)
	Thumb any
	// TTL is self-destruct timer in seconds (for secret media)
	TTL int32
	// Spoiler hides media behind spoiler overlay
	Spoiler bool
	// WithoutReply sends media without replying to the original message
	WithoutReply bool
	// ReplyMessageID is the message ID to reply to (instead of the current message)
	ReplyMessageID int
}

// Reply sends a reply to the current update's message.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
func (u *Update) Reply(Text any, Opts ...*ReplyOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, mtp_errors.ErrTextEmpty
	}

	var opts *ReplyOpts
	if len(Opts) > 0 && Opts[0] != nil {
		opts = Opts[0]
	} else {
		opts = &ReplyOpts{}
	}

	// Default parse mode is HTML
	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Get chat ID
	chatID := u.ChatID()

	// Convert text to string
	var message string
	switch v := Text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

	// Parse HTML/Markdown to entities (case-insensitive)
	var text string
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

		result, err := parsemode.Parse(message, mode)
		if err == nil && result != nil {
			text = result.Text
			entities = result.Entities
		} else {
			text = message
		}
	} else {
		text = message
	}

	// Build the send request
	req := &tg.MessagesSendMessageRequest{
		Message:  text,
		Entities: entities,
	}

	// Set reply to message ID
	// Priority: ReplyMessageID > ReplyTo field > current message ID
	if opts.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opts.ReplyMessageID,
		}
	} else if opts.ReplyTo != nil {
		req.ReplyTo = opts.ReplyTo
	} else if u.EffectiveMessage != nil && !opts.WithoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: u.EffectiveMessage.ID,
		}
	}

	// Set reply markup
	if opts.Markup != nil {
		req.ReplyMarkup = opts.Markup
	}

	// Set flags
	if opts.NoWebpage {
		req.Flags |= 4 // NoWebpage flag
		req.NoWebpage = true
	}
	if opts.Silent {
		req.Silent = true
	}
	if opts.ClearDraft {
		req.ClearDraft = true
	}
	if opts.LinkPreview {
		// Enable link preview
	}
	if opts.NoForwards {
		req.Noforwards = true
	}
	if opts.TopicID != 0 {
		// Set topic ID for forum posts
	}
	if opts.ScheduleDate != 0 {
		req.ScheduleDate = int(opts.ScheduleDate)
	}
	if opts.SendAs != 0 {
		// Set send as peer
	}
	if opts.TTL != 0 {
		// Set TTL
	}
	if opts.Effect != 0 {
		req.Effect = opts.Effect
	}

	return u.Ctx.SendMessage(chatID, req)
}

// ReplyMedia sends a media reply to the current update's message.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
// Default parse mode for caption is HTML.
//
// Example using InputMedia:
//
//	newMsg, err := u.ReplyMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, &adapter.ReplyMediaOpts{
//	    Caption: "Photo caption",
//	})
//
// For using fileID strings, see ReplyMediaWithFileID.
func (u *Update) ReplyMedia(Media tg.InputMediaClass, Opts ...*ReplyMediaOpts) (*types.Message, error) {
	if Media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}

	var opts *ReplyMediaOpts
	if len(Opts) > 0 && Opts[0] != nil {
		opts = Opts[0]
	} else {
		opts = &ReplyMediaOpts{}
	}

	// Default parse mode is HTML
	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Get chat ID
	chatID := u.ChatID()

	// Parse caption for entities
	var caption string
	var entities []tg.MessageEntityClass

	if opts.Caption != "" && parseMode != ModeNone {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case HTML:
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(opts.Caption, mode)
		if err == nil && result != nil {
			caption = result.Text
			entities = result.Entities
		} else {
			caption = opts.Caption
		}
	} else {
		caption = opts.Caption
	}

	// Build the send media request
	req := &tg.MessagesSendMediaRequest{
		Media:    Media,
		Message:  caption,
		Entities: entities,
	}

	// Set reply to message ID
	// Priority: ReplyMessageID > ReplyTo field > current message ID
	if opts.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opts.ReplyMessageID,
		}
	} else if opts.ReplyTo != nil {
		req.ReplyTo = opts.ReplyTo
	} else if u.EffectiveMessage != nil && !opts.WithoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: u.EffectiveMessage.ID,
		}
	}

	// Set reply markup
	if opts.Markup != nil {
		req.ReplyMarkup = opts.Markup
	}

	// Set flags
	if opts.Silent {
		req.Silent = true
	}
	if opts.ClearDraft {
		req.ClearDraft = true
	}
	if opts.LinkPreview {
		// Enable link preview
	}
	if opts.NoForwards {
		req.Noforwards = true
	}
	if opts.TopicID != 0 {
		// Set topic ID for forum posts
	}
	if opts.ScheduleDate != 0 {
		req.ScheduleDate = int(opts.ScheduleDate)
	}
	if opts.SendAs != 0 {
		// Set send as peer
	}
	if opts.TTL != 0 {
		// Set TTL
	}
	if opts.InvertMedia {
		req.InvertMedia = true
	}

	return u.Ctx.SendMedia(chatID, req)
}

// ReplyMediaWithFileID sends a media reply to the current update's message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
// Default parse mode for caption is HTML.
//
// Example:
//
//	fileID := msg.FileID()  // or msg.Document().FileID() or msg.Photo().FileID()
//	newMsg, err := u.ReplyMediaWithFileID(fileID, &adapter.ReplyMediaOpts{
//	    Caption: "Here's the media you requested",
//	})
func (u *Update) ReplyMediaWithFileID(fileID string, Opts ...*ReplyMediaOpts) (*types.Message, error) {
	// Extract caption from opts to pass to InputMediaFromFileID
	var caption string
	if len(Opts) > 0 && Opts[0] != nil {
		caption = Opts[0].Caption
	}

	media, err := types.InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return u.ReplyMedia(media, Opts...)
}

// Edit edits the current update's message text.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
// For callback queries, edits the message that triggered the callback.
func (u *Update) Edit(Text any, Opts ...*ReplyOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, mtp_errors.ErrTextEmpty
	}

	// Handle callback queries - edit the message that triggered the callback
	if u.EffectiveMessage == nil && u.CallbackQuery != nil {
		var opts *ReplyOpts
		if len(Opts) > 0 && Opts[0] != nil {
			opts = Opts[0]
		} else {
			opts = &ReplyOpts{}
		}

		// Default parse mode is HTML
		parseMode := opts.ParseMode
		if parseMode == "" {
			parseMode = HTML
		}

		// Convert text to string
		var message string
		switch v := Text.(type) {
		case string:
			message = v
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			message = fmt.Sprintf("%v", v)
		default:
			message = fmt.Sprintf("%v", v)
		}

		// Parse HTML/Markdown to entities
		var text string
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

			result, err := parsemode.Parse(message, mode)
			if err == nil && result != nil {
				text = result.Text
				entities = result.Entities
			} else {
				text = message
			}
		} else {
			text = message
		}

		// Build the edit request for callback query
		req := &tg.MessagesEditMessageRequest{
			ID:        u.MsgID(),
			Message:   text,
			Entities:  entities,
			NoWebpage: opts.NoWebpage,
		}

		// Set reply markup
		if opts.Markup != nil {
			req.ReplyMarkup = opts.Markup
		}

		return u.Ctx.EditMessage(u.ChatID(), req)
	}

	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	var opts *ReplyOpts
	if len(Opts) > 0 && Opts[0] != nil {
		opts = Opts[0]
	} else {
		opts = &ReplyOpts{}
	}

	// Default parse mode is HTML
	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Convert text to string
	var message string
	switch v := Text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

	// Parse HTML/Markdown to entities
	var text string
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

		result, err := parsemode.Parse(message, mode)
		if err == nil && result != nil {
			text = result.Text
			entities = result.Entities
		} else {
			text = message
		}
	} else {
		text = message
	}

	// Build the edit request
	req := &tg.MessagesEditMessageRequest{
		ID:        u.EffectiveMessage.ID,
		Message:   text,
		Entities:  entities,
		NoWebpage: opts.NoWebpage,
	}

	// Set reply markup
	if opts.Markup != nil {
		req.ReplyMarkup = opts.Markup
	}

	return u.Ctx.EditMessage(u.ChatID(), req)
}

// EditMedia edits the media of the current update's message.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
//
// Example using InputMedia:
//
//	editedMsg, err := u.EditMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, &adapter.ReplyMediaOpts{
//	    Caption: "New photo",
//	})
//
// For using fileID strings, see EditMediaWithFileID.
func (u *Update) EditMedia(Media tg.InputMediaClass, Opts ...*ReplyMediaOpts) (*types.Message, error) {
	if Media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}
	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	var opts *ReplyMediaOpts
	if len(Opts) > 0 && Opts[0] != nil {
		opts = Opts[0]
	} else {
		opts = &ReplyMediaOpts{}
	}

	// Default parse mode is HTML
	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Parse caption for entities
	var caption string
	var entities []tg.MessageEntityClass

	if opts.Caption != "" && parseMode != ModeNone {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case HTML:
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(opts.Caption, mode)
		if err == nil && result != nil {
			caption = result.Text
			entities = result.Entities
		} else {
			caption = opts.Caption
		}
	} else {
		caption = opts.Caption
	}

	// Build the edit request
	req := &tg.MessagesEditMessageRequest{
		ID:       u.EffectiveMessage.ID,
		Media:    Media,
		Message:  caption,
		Entities: entities,
	}

	// Set reply markup
	if opts.Markup != nil {
		req.ReplyMarkup = opts.Markup
	}

	return u.Ctx.EditMessage(u.ChatID(), req)
}

// EditMediaWithFileID edits the media of the current update's message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
//
// Example:
//
//	fileID := newMsg.FileID()
//	editedMsg, err := u.EditMediaWithFileID(fileID, &adapter.ReplyMediaOpts{
//	    Caption: "Updated media!",
//	})
func (u *Update) EditMediaWithFileID(fileID string, Opts ...*ReplyMediaOpts) (*types.Message, error) {
	// Extract caption from opts to pass to InputMediaFromFileID
	var caption string
	if len(Opts) > 0 && Opts[0] != nil {
		caption = Opts[0].Caption
	}

	media, err := types.InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return u.EditMedia(media, Opts...)
}

// EditCaption edits the caption of the current update's media message.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
func (u *Update) EditCaption(Text any, Opts ...*ReplyOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, mtp_errors.ErrTextEmpty
	}
	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	var opts *ReplyOpts
	if len(Opts) > 0 && Opts[0] != nil {
		opts = Opts[0]
	} else {
		opts = &ReplyOpts{}
	}

	// Default parse mode is HTML
	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	// Convert text to string
	var message string
	switch v := Text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

	// Parse HTML/Markdown to entities
	var text string
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

		result, err := parsemode.Parse(message, mode)
		if err == nil && result != nil {
			text = result.Text
			entities = result.Entities
		} else {
			text = message
		}
	} else {
		text = message
	}

	// Build the edit request (keep existing media, just update caption)
	req := &tg.MessagesEditMessageRequest{
		ID:       u.EffectiveMessage.ID,
		Message:  text,
		Entities: entities,
	}

	// Set reply markup
	if opts.Markup != nil {
		req.ReplyMarkup = opts.Markup
	}

	return u.Ctx.EditMessage(u.ChatID(), req)
}

// EditReplyMarkup edits only the reply markup of the current update's message.
func (u *Update) EditReplyMarkup(Markup tg.ReplyMarkupClass) (*types.Message, error) {
	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	req := &tg.MessagesEditMessageRequest{
		ID:          u.EffectiveMessage.ID,
		ReplyMarkup: Markup,
	}

	return u.Ctx.EditMessage(u.ChatID(), req)
}

// ChatType returns the type of chat for this update.
// Returns "private" for users, "group" for chats, "channel" for channels,
// or an empty string if the chat type cannot be determined.
func (u *Update) ChatType() string {
	chat := u.EffectiveChat()
	if chat == nil {
		return ""
	}

	switch chat.(type) {
	case *types.User:
		return "private"
	case *types.Chat:
		return "group"
	case *types.Channel:
		return "channel"
	}
	return ""
}
