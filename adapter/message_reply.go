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

// Reply sends a reply to the current update's message.
// Text can be a string or any type that can be formatted with %v.
// Uses the client's default parse mode from ClientOpts.ParseMode.
func (u *Update) Reply(text any, opts ...*SendOpts) (*types.Message, error) {
	if text == "" || text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	opt := functions.GetOptDef(&SendOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	chatID := u.ChatID()

	var message string
	switch v := text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

	var finalText string
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

		result, err := parsemode.Parse(message, mode)
		if err == nil && result != nil {
			finalText = result.Text
			entities = result.Entities
		} else {
			finalText = message
		}
	} else {
		finalText = message
	}

	if len(opt.Entities) > 0 {
		entities = opt.Entities
	}

	req := &tg.MessagesSendMessageRequest{
		Message:                finalText,
		Entities:               entities,
		NoWebpage:              opt.NoWebpage,
		Silent:                 opt.Silent,
		Background:             opt.Background,
		ClearDraft:             opt.ClearDraft,
		Noforwards:             opt.Noforwards,
		UpdateStickersetsOrder: opt.UpdateStickersetsOrder,
		InvertMedia:            opt.InvertMedia,
		AllowPaidFloodskip:     opt.AllowPaidFloodskip,
	}

	if opt.ReplyMarkup != nil {
		req.ReplyMarkup = opt.ReplyMarkup
	}
	if opt.Peer != nil {
		req.Peer = opt.Peer
	}
	if opt.ReplyTo != nil {
		req.ReplyTo = opt.ReplyTo
	}
	if opt.RandomID != 0 {
		req.RandomID = opt.RandomID
	}
	if opt.ScheduleDate != 0 {
		req.ScheduleDate = opt.ScheduleDate
	}
	if opt.ScheduleRepeatPeriod != 0 {
		req.ScheduleRepeatPeriod = opt.ScheduleRepeatPeriod
	}
	if opt.SendAs != nil {
		req.SendAs = opt.SendAs
	}
	if opt.QuickReplyShortcut != nil {
		req.QuickReplyShortcut = opt.QuickReplyShortcut
	}
	if opt.Effect != 0 {
		req.Effect = opt.Effect
	}
	if opt.AllowPaidStars != 0 {
		req.AllowPaidStars = opt.AllowPaidStars
	}
	if opt.SuggestedPost != (tg.SuggestedPost{}) {
		req.SuggestedPost = opt.SuggestedPost
	}

	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	} else if req.ReplyTo == nil && u.HasMessage() && !opt.WithoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: u.EffectiveMessage.ID,
		}
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.SendMessage(chatID, req, connID)
}

// ReplyMedia sends a media reply to the current update's message.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
// Uses the client's default parse mode from ClientOpts.ParseMode for caption formatting.
//
// Example using InputMedia:
//
//	newMsg, err := u.ReplyMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, &adapter.SendMediaOpts{
//	    Caption: "Photo caption",
//	})
//
// For using fileID strings, see ReplyMediaWithFileID.
func (u *Update) ReplyMedia(media tg.InputMediaClass, opts ...*SendMediaOpts) (*types.Message, error) {
	if media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}

	opt := functions.GetOptDef(&SendMediaOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	chatID := u.ChatID()

	var caption string
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
			caption = result.Text
			entities = result.Entities
		} else {
			caption = opt.Caption
		}
	} else {
		caption = opt.Caption
	}

	if len(opt.Entities) > 0 {
		entities = opt.Entities
	}

	req := &tg.MessagesSendMediaRequest{
		Media:                  media,
		Message:                caption,
		Entities:               entities,
		Silent:                 opt.Silent,
		Background:             opt.Background,
		ClearDraft:             opt.ClearDraft,
		Noforwards:             opt.Noforwards,
		UpdateStickersetsOrder: opt.UpdateStickersetsOrder,
		InvertMedia:            opt.InvertMedia,
		AllowPaidFloodskip:     opt.AllowPaidFloodskip,
	}

	if opt.ReplyMarkup != nil {
		req.ReplyMarkup = opt.ReplyMarkup
	}
	if opt.Peer != nil {
		req.Peer = opt.Peer
	}
	if opt.ReplyTo != nil {
		req.ReplyTo = opt.ReplyTo
	}
	if opt.RandomID != 0 {
		req.RandomID = opt.RandomID
	}
	if opt.ScheduleDate != 0 {
		req.ScheduleDate = opt.ScheduleDate
	}
	if opt.SendAs != nil {
		req.SendAs = opt.SendAs
	}
	if opt.QuickReplyShortcut != nil {
		req.QuickReplyShortcut = opt.QuickReplyShortcut
	}
	if opt.Effect != 0 {
		req.Effect = opt.Effect
	}
	if opt.AllowPaidStars != 0 {
		req.AllowPaidStars = opt.AllowPaidStars
	}
	if opt.SuggestedPost != (tg.SuggestedPost{}) {
		req.SuggestedPost = opt.SuggestedPost
	}

	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	} else if req.ReplyTo == nil && u.HasMessage() && !opt.WithoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: u.EffectiveMessage.ID,
		}
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.SendMedia(chatID, req, connID)
}

// ReplyMediaWithFileID sends a media reply to the current update's message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
// Default parse mode for caption is HTML.
//
// Example:
//
//	fileID := msg.FileID()  // or msg.Document().FileID() or msg.Photo().FileID()
//	newMsg, err := u.ReplyMediaWithFileID(fileID, &adapter.SendMediaOpts{
//	    Caption: "Here's the media you requested",
//	})
func (u *Update) ReplyMediaWithFileID(fileID string, opts ...*SendMediaOpts) (*types.Message, error) {
	var caption string
	if len(opts) > 0 && opts[0] != nil {
		caption = opts[0].Caption
	}

	media, err := types.InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return u.ReplyMedia(media, opts...)
}
