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

// Edit edits the current update's message text.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
// For callback queries, edits the message that triggered the callback.
func (u *Update) Edit(Text any, Opts ...*SendOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	if u.EffectiveMessage == nil && u.CallbackQuery != nil {
		opts := functions.GetOptDef(&SendOpts{}, Opts...)

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
			ID:        u.MsgID(),
			Message:   text,
			Entities:  entities,
			NoWebpage: opts.NoWebpage,
		}

		if opts.ReplyMarkup != nil {
			req.ReplyMarkup = opts.ReplyMarkup
		}

		return u.Ctx.EditMessage(u.ChatID(), req)
	}

	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	opts := functions.GetOptDef(&SendOpts{}, Opts...)

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
		ID:        u.EffectiveMessage.ID,
		Message:   text,
		Entities:  entities,
		NoWebpage: opts.NoWebpage,
	}

	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
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
//	}, &adapter.SendMediaOpts{
//	    Caption: "New photo",
//	})
//
// For using fileID strings, see EditMediaWithFileID.
func (u *Update) EditMedia(Media tg.InputMediaClass, Opts ...*SendMediaOpts) (*types.Message, error) {
	if Media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}
	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	opts := functions.GetOptDef(&SendMediaOpts{}, Opts...)

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
		ID:       u.EffectiveMessage.ID,
		Media:    Media,
		Message:  caption,
		Entities: entities,
	}

	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
	}

	return u.Ctx.EditMessage(u.ChatID(), req)
}

// EditMediaWithFileID edits the media of the current update's message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
//
// Example:
//
//	fileID := newMsg.FileID()
//	editedMsg, err := u.EditMediaWithFileID(fileID, &adapter.SendMediaOpts{
//	    Caption: "Updated media!",
//	})
func (u *Update) EditMediaWithFileID(fileID string, Opts ...*SendMediaOpts) (*types.Message, error) {
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
func (u *Update) EditCaption(Text any, Opts ...*SendOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}
	if u.EffectiveMessage == nil {
		return nil, fmt.Errorf("no effective message to edit")
	}

	opts := functions.GetOptDef(&SendOpts{}, Opts...)

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
		ID:       u.EffectiveMessage.ID,
		Message:  text,
		Entities: entities,
	}

	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
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
