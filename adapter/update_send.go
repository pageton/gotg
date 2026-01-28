package adapter

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/types"
)

// SendMessage sends a text message to the specified chat.
// Text can be a string or any type that can be formatted with %v.
// Default parse mode is HTML.
//
// NOTE: This method does NOT reply by default. To reply to a specific message,
// set ReplyMessageID in SendOpts.
//
// Parameters:
//   - chatID: The target chat ID (use 0 to use the current update's chat)
//   - text: The message text
//   - opts: Optional SendOpts for formatting, reply markup, etc.
//
// Returns the sent Message or an error.
//
// Example:
//
//	msg, err := u.SendMessage(0, "Hello, world!")  // Send to current chat (no reply)
//	msg, err := u.SendMessage(chatID, "<b>Bold text</b>", &SendOpts{
//	    ParseMode: "HTML",
//	})
//	// Reply to a specific message:
//	msg, err := u.SendMessage(chatID, "Reply text", &SendOpts{
//	    ReplyMessageID: 123,
//	})
func (u *Update) SendMessage(chatID int64, text string, opts ...*SendOpts) (*types.Message, error) {
	if chatID == 0 {
		chatID = u.ChatID()
	}
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	// Build request from opts
	opt := functions.GetOptDef(&SendOpts{}, opts...)

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

	if len(opt.Entities) > 0 {
		entities = opt.Entities
	}

	req := &tg.MessagesSendMessageRequest{
		Message:                messageText,
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

	// Add ReplyTo if ReplyMessageID is set
	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.SendMessage(chatID, req, connID)
}

// SendMedia sends media (photo, document, video, etc.) to the chat associated with this update.
// Accepts tg.InputMediaClass (e.g., InputMediaPhoto, InputMediaDocument).
// Default parse mode for caption is HTML.
//
// NOTE: This method does NOT reply by default. To reply to a specific message,
// set ReplyMessageID in SendMediaOpts.
//
// Parameters:
//   - media: The media to send (tg.InputMediaClass)
//   - caption: Optional caption text
//   - opts: Optional SendMediaOpts
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
//	msg, err := u.SendMedia(media, "caption", &SendMediaOpts{
//	    ReplyMessageID: 123,
//	})
func (u *Update) SendMedia(media tg.InputMediaClass, caption string, opts ...*SendMediaOpts) (*types.Message, error) {
	chatID := u.ChatID()
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	opt := functions.GetOptDef(&SendMediaOpts{}, opts...)

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

	if len(opt.Entities) > 0 {
		entities = opt.Entities
	}

	req := &tg.MessagesSendMediaRequest{
		Media:                  media,
		Message:                captionText,
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
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.SendMedia(chatID, req, connID)
}

// SendMultiMedia sends multiple media items as an album to the chat associated with this update.
// Albums can contain up to 10 media items.
//
// NOTE: This method does NOT reply by default. To reply to a specific message,
// set ReplyMessageID in SendMediaOpts.
//
// Parameters:
//   - media: Slice of InputMediaClass items
//   - opts: Optional SendMediaOpts (applied to all items)
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
//	msgs, err := u.SendMultiMedia(media, &SendMediaOpts{
//	    ReplyMessageID: 123,
//	})
func (u *Update) SendMultiMedia(media []tg.InputMediaClass, opts ...*SendMediaOpts) (*types.Message, error) {
	chatID := u.ChatID()
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}

	if len(media) == 0 {
		return nil, fmt.Errorf("media slice cannot be empty")
	}

	opt := functions.GetOptDef(&SendMediaOpts{}, opts...)

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
		MultiMedia:             inputMedia,
		Silent:                 opt.Silent,
		Background:             opt.Background,
		ClearDraft:             opt.ClearDraft,
		Noforwards:             opt.Noforwards,
		UpdateStickersetsOrder: opt.UpdateStickersetsOrder,
		InvertMedia:            opt.InvertMedia,
		AllowPaidFloodskip:     opt.AllowPaidFloodskip,
		ScheduleDate:           opt.ScheduleDate,
	}

	if opt.Peer != nil {
		req.Peer = opt.Peer
	}
	if opt.ReplyTo != nil {
		req.ReplyTo = opt.ReplyTo
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

	if opt.ReplyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: opt.ReplyMessageID,
		}
	}

	connID := opt.BusinessConnectionID
	if connID == "" {
		connID = u.ConnectionID()
	}
	return u.Ctx.SendMultiMedia(chatID, req, connID)
}
