package adapter

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/types"
)

// EditMessage edits a message in the specified chat.
// Text can be a string or any type that can be formatted with %v.
// Uses the client's default parse mode from ClientOpts.ParseMode.
//
// Parameters:
//   - chatID: The target chat ID (use 0 to use the current update's chat)
//   - messageID: The ID of the message to edit
//   - text: New message text
//   - opts: Optional EditOpts for entities, reply markup, etc.
//
// Returns the edited Message or an error.
//
// Example:
//
//	msg, err := u.EditMessage(0, 123, "Updated text")  // Edit in current chat
//	msg, err := u.EditMessage(chatID, 123, "<b>Updated</b>", &EditOpts{
//	    ParseMode: "HTML",
//	})
func (u *Update) EditMessage(chatID int64, messageID int, text string, opts ...*EditOpts) (*types.Message, error) {
	if chatID == 0 {
		chatID = u.ChatID()
	}
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	opt := functions.GetOptDef(&EditOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	var messageText string
	var entities []tg.MessageEntityClass

	if parseMode != "" && parseMode != ModeNone {
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

	req := &tg.MessagesEditMessageRequest{
		ID:                   messageID,
		Message:              messageText,
		Entities:             entities,
		NoWebpage:            opt.NoWebpage,
		InvertMedia:          opt.InvertMedia,
		ScheduleDate:         opt.ScheduleDate,
		ScheduleRepeatPeriod: opt.ScheduleRepeatPeriod,
		QuickReplyShortcutID: opt.QuickReplyShortcutID,
	}

	if opt.Peer != nil {
		req.Peer = opt.Peer
	}
	if opt.ID != 0 {
		req.ID = opt.ID
	}
	if opt.Media != nil {
		req.Media = opt.Media
	}
	if opt.ReplyMarkup != nil {
		req.ReplyMarkup = opt.ReplyMarkup
	}
	if len(opt.Entities) > 0 {
		req.Entities = opt.Entities
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.EditMessage(chatID, req, connID)
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
//   - opts: Optional EditMediaOpts
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
func (u *Update) EditMessageMedia(chatID int64, messageID int, media tg.InputMediaClass, caption string, opts ...*EditMediaOpts) (*types.Message, error) {
	if chatID == 0 {
		chatID = u.ChatID()
	}
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	opt := functions.GetOptDef(&EditMediaOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	if caption != "" && opt.Caption == "" {
		opt.Caption = caption
	}

	var captionText string
	var entities []tg.MessageEntityClass

	if opt.Caption != "" && parseMode != "" && parseMode != ModeNone {
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

	req := &tg.MessagesEditMessageRequest{
		ID:                   messageID,
		Media:                media,
		Message:              captionText,
		Entities:             entities,
		NoWebpage:            opt.NoWebpage,
		InvertMedia:          opt.InvertMedia,
		ScheduleDate:         opt.ScheduleDate,
		ScheduleRepeatPeriod: opt.ScheduleRepeatPeriod,
		QuickReplyShortcutID: opt.QuickReplyShortcutID,
	}

	if opt.Peer != nil {
		req.Peer = opt.Peer
	}
	if opt.ID != 0 {
		req.ID = opt.ID
	}
	if opt.Media != nil {
		req.Media = opt.Media
	}
	if opt.ReplyMarkup != nil {
		req.ReplyMarkup = opt.ReplyMarkup
	}
	if len(opt.Entities) > 0 {
		req.Entities = opt.Entities
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.EditMessage(chatID, req, connID)
}
