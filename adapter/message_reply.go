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
func (u *Update) Reply(Text any, Opts ...*SendOpts) (*types.Message, error) {
	if Text == "" || Text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	opts := functions.GetOptDef(&SendOpts{}, Opts...)

	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	chatID := u.ChatID()

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
			text = result.Text
			entities = result.Entities
		} else {
			text = message
		}
	} else {
		text = message
	}

	if len(opts.Entities) > 0 {
		entities = opts.Entities
	}

	req := &tg.MessagesSendMessageRequest{
		Message:                text,
		Entities:               entities,
		NoWebpage:              opts.NoWebpage,
		Silent:                 opts.Silent,
		Background:             opts.Background,
		ClearDraft:             opts.ClearDraft,
		Noforwards:             opts.Noforwards,
		UpdateStickersetsOrder: opts.UpdateStickersetsOrder,
		InvertMedia:            opts.InvertMedia,
		AllowPaidFloodskip:     opts.AllowPaidFloodskip,
	}

	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
	}
	if opts.Peer != nil {
		req.Peer = opts.Peer
	}
	if opts.ReplyTo != nil {
		req.ReplyTo = opts.ReplyTo
	}
	if opts.RandomID != 0 {
		req.RandomID = opts.RandomID
	}
	if opts.ScheduleDate != 0 {
		req.ScheduleDate = opts.ScheduleDate
	}
	if opts.ScheduleRepeatPeriod != 0 {
		req.ScheduleRepeatPeriod = opts.ScheduleRepeatPeriod
	}
	if opts.SendAs != nil {
		req.SendAs = opts.SendAs
	}
	if opts.QuickReplyShortcut != nil {
		req.QuickReplyShortcut = opts.QuickReplyShortcut
	}
	if opts.Effect != 0 {
		req.Effect = opts.Effect
	}
	if opts.AllowPaidStars != 0 {
		req.AllowPaidStars = opts.AllowPaidStars
	}
	if opts.SuggestedPost != (tg.SuggestedPost{}) {
		req.SuggestedPost = opts.SuggestedPost
	}

	if opts.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opts.ReplyMessageID,
		}
	} else if req.ReplyTo == nil && u.HasMessage() && !opts.WithoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: u.EffectiveMessage.ID,
		}
	}

	connID := opts.BusinessConnectionID
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
func (u *Update) ReplyMedia(Media tg.InputMediaClass, Opts ...*SendMediaOpts) (*types.Message, error) {
	if Media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}

	opts := functions.GetOptDef(&SendMediaOpts{}, Opts...)

	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	chatID := u.ChatID()

	var caption string
	var entities []tg.MessageEntityClass

	if opts.Caption != "" && parseMode != "" && parseMode != ModeNone {
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

	if len(opts.Entities) > 0 {
		entities = opts.Entities
	}

	req := &tg.MessagesSendMediaRequest{
		Media:                  Media,
		Message:                caption,
		Entities:               entities,
		Silent:                 opts.Silent,
		Background:             opts.Background,
		ClearDraft:             opts.ClearDraft,
		Noforwards:             opts.Noforwards,
		UpdateStickersetsOrder: opts.UpdateStickersetsOrder,
		InvertMedia:            opts.InvertMedia,
		AllowPaidFloodskip:     opts.AllowPaidFloodskip,
	}

	if opts.ReplyMarkup != nil {
		req.ReplyMarkup = opts.ReplyMarkup
	}
	if opts.Peer != nil {
		req.Peer = opts.Peer
	}
	if opts.ReplyTo != nil {
		req.ReplyTo = opts.ReplyTo
	}
	if opts.RandomID != 0 {
		req.RandomID = opts.RandomID
	}
	if opts.ScheduleDate != 0 {
		req.ScheduleDate = opts.ScheduleDate
	}
	if opts.SendAs != nil {
		req.SendAs = opts.SendAs
	}
	if opts.QuickReplyShortcut != nil {
		req.QuickReplyShortcut = opts.QuickReplyShortcut
	}
	if opts.Effect != 0 {
		req.Effect = opts.Effect
	}
	if opts.AllowPaidStars != 0 {
		req.AllowPaidStars = opts.AllowPaidStars
	}
	if opts.SuggestedPost != (tg.SuggestedPost{}) {
		req.SuggestedPost = opts.SuggestedPost
	}

	if opts.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opts.ReplyMessageID,
		}
	} else if req.ReplyTo == nil && u.HasMessage() && !opts.WithoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: u.EffectiveMessage.ID,
		}
	}

	connID := opts.BusinessConnectionID
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
func (u *Update) ReplyMediaWithFileID(fileID string, Opts ...*SendMediaOpts) (*types.Message, error) {
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
