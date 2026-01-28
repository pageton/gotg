package adapter

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	gotgErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/types"
)

// getEditMessageID returns the message ID for editing operations.
// Supports both regular messages and callback queries.
func (u *Update) getEditMessageID() (int, error) {
	if u.HasMessage() {
		return u.EffectiveMessage.ID, nil
	}
	if u.CallbackQuery != nil {
		return u.MsgID(), nil
	}
	return 0, fmt.Errorf("no effective message to edit")
}

// Edit edits the current update's message text.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
// For callback queries, edits the message that triggered the callback.
func (u *Update) Edit(Text any, Opts ...*EditOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	opts := functions.GetOptDef(&EditOpts{}, Opts...)

	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	var message string
	switch v := Text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

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

	req := &tg.MessagesEditMessageRequest{
		ID:                   msgID,
		Message:              text,
		Entities:             entities,
		NoWebpage:            opts.NoWebpage,
		InvertMedia:          opts.InvertMedia,
		ScheduleDate:         opts.ScheduleDate,
		ScheduleRepeatPeriod: opts.ScheduleRepeatPeriod,
		QuickReplyShortcutID: opts.QuickReplyShortcutID,
	}

	if opts.Peer != nil {
		req.Peer = opts.Peer
	}
	if opts.ID != 0 {
		req.ID = opts.ID
	}
	if opts.Media != nil {
		req.Media = opts.Media
	}
	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
	}
	if len(opts.Entities) > 0 {
		req.Entities = opts.Entities
	}

	connID := opts.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.EditMessage(u.ChatID(), req, connID)
}

// EditMedia edits the media of the current update's message.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
//
// Example using InputMedia:
//
//	editedMsg, err := u.EditMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, &adapter.EditMediaOpts{
//	    Caption: "New photo",
//	})
//
// For using fileID strings, see EditMediaWithFileID.
func (u *Update) EditMedia(Media tg.InputMediaClass, Opts ...*EditMediaOpts) (*types.Message, error) {
	if Media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	opts := functions.GetOptDef(&EditMediaOpts{}, Opts...)

	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

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

	req := &tg.MessagesEditMessageRequest{
		ID:                   msgID,
		Media:                Media,
		Message:              caption,
		Entities:             entities,
		NoWebpage:            opts.NoWebpage,
		InvertMedia:          opts.InvertMedia,
		ScheduleDate:         opts.ScheduleDate,
		ScheduleRepeatPeriod: opts.ScheduleRepeatPeriod,
		QuickReplyShortcutID: opts.QuickReplyShortcutID,
	}

	if opts.Peer != nil {
		req.Peer = opts.Peer
	}
	if opts.ID != 0 {
		req.ID = opts.ID
	}
	if opts.Media != nil {
		req.Media = opts.Media
	}
	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
	}
	if len(opts.Entities) > 0 {
		req.Entities = opts.Entities
	}

	connID := opts.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.EditMessage(u.ChatID(), req, connID)
}

// EditMediaWithFileID edits the media of the current update's message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
//
// Example:
//
//	fileID := newMsg.FileID()
//	editedMsg, err := u.EditMediaWithFileID(fileID, &adapter.EditMediaOpts{
//	    Caption: "Updated media!",
//	})
func (u *Update) EditMediaWithFileID(fileID string, Opts ...*EditMediaOpts) (*types.Message, error) {
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
func (u *Update) EditCaption(Text any, Opts ...*EditOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	opts := functions.GetOptDef(&EditOpts{}, Opts...)

	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = HTML
	}

	var message string
	switch v := Text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

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

	req := &tg.MessagesEditMessageRequest{
		ID:                   msgID,
		Message:              text,
		Entities:             entities,
		NoWebpage:            opts.NoWebpage,
		InvertMedia:          opts.InvertMedia,
		ScheduleDate:         opts.ScheduleDate,
		ScheduleRepeatPeriod: opts.ScheduleRepeatPeriod,
		QuickReplyShortcutID: opts.QuickReplyShortcutID,
	}

	if opts.Peer != nil {
		req.Peer = opts.Peer
	}
	if opts.ID != 0 {
		req.ID = opts.ID
	}
	if opts.Media != nil {
		req.Media = opts.Media
	}
	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
	}
	if len(opts.Entities) > 0 {
		req.Entities = opts.Entities
	}

	connID := opts.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.EditMessage(u.ChatID(), req, connID)
}

// EditReplyMarkup edits only the reply markup of the current update's message.
// For callback queries, edits the message that triggered the callback.
func (u *Update) EditReplyMarkup(Markup tg.ReplyMarkupClass) (*types.Message, error) {
	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	req := &tg.MessagesEditMessageRequest{
		ID:          msgID,
		ReplyMarkup: Markup,
	}

	return u.Ctx.EditMessage(u.ChatID(), req, u.ConnectionID())
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
