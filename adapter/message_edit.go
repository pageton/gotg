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
// Supports regular messages, callback queries, and inline callback queries.
func (u *Update) getEditMessageID() (int, error) {
	if u.HasMessage() {
		return u.EffectiveMessage.ID, nil
	}
	if u.CallbackQuery != nil {
		return u.MsgID(), nil
	}
	if u.InlineCallbackQuery != nil {
		// Inline messages don't have integer IDs; callers should
		// check isInlineCallback() and use the inline edit path instead.
		return 0, nil
	}
	return 0, fmt.Errorf("no effective message to edit")
}

// isInlineCallback returns true when the current update is an inline message callback.
func (u *Update) isInlineCallback() bool {
	return u.InlineCallbackQuery != nil
}

// isChosenInlineResult returns true when the current update is a chosen inline result
// that has an inline message ID available for editing.
func (u *Update) isChosenInlineResult() bool {
	return u.ChosenInlineResult != nil && u.ChosenInlineResult.HasInlineMessageID()
}

// isInlineEdit returns true when the current update can be edited via inline message ID,
// either from an inline callback query or a chosen inline result.
func (u *Update) isInlineEdit() bool {
	return u.isInlineCallback() || u.isChosenInlineResult()
}

// getInlineMessageID returns the inline message ID from either an inline callback query
// or a chosen inline result. Returns nil if neither is available.
func (u *Update) getInlineMessageID() tg.InputBotInlineMessageIDClass {
	if u.InlineCallbackQuery != nil {
		return u.InlineCallbackQuery.MsgID
	}
	if u.ChosenInlineResult != nil {
		return u.ChosenInlineResult.GetInlineMessageID()
	}
	return nil
}

// invokeInlineEdit sends a MessagesEditInlineBotMessageRequest to the correct DC
// extracted from the inline message ID. Telegram requires inline edits to be
// sent to the DC that owns the inline message.
func (u *Update) invokeInlineEdit(req *tg.MessagesEditInlineBotMessageRequest) error {
	dcID := req.ID.GetDCID()

	if dcID != 0 && u.Ctx.GetDCPool != nil {
		invoker, err := u.Ctx.GetDCPool(u.Ctx, dcID)
		if err == nil {
			dcClient := tg.NewClient(invoker)
			_, err = dcClient.MessagesEditInlineBotMessage(u.Ctx, req)
			return err
		}
	}

	_, err := u.Ctx.Raw.MessagesEditInlineBotMessage(u.Ctx, req)
	return err
}

func (u *Update) editInlineText(text string, entities []tg.MessageEntityClass, opts *EditOpts) (*types.Message, error) {
	inlineMsgID := u.getInlineMessageID()
	if inlineMsgID == nil {
		return nil, fmt.Errorf("no inline message ID available for editing")
	}
	req := &tg.MessagesEditInlineBotMessageRequest{
		ID:          inlineMsgID,
		NoWebpage:   opts.NoWebpage,
		InvertMedia: opts.InvertMedia,
	}
	req.SetMessage(text)
	if len(entities) > 0 {
		req.SetEntities(entities)
	}
	if opts.ReplyMarkup != nil {
		req.SetReplyMarkup(opts.ReplyMarkup)
	}
	if opts.Media != nil {
		req.SetMedia(opts.Media)
	}
	if err := u.invokeInlineEdit(req); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *Update) editInlineMedia(media tg.InputMediaClass, caption string, entities []tg.MessageEntityClass, opts *EditMediaOpts) (*types.Message, error) {
	inlineMsgID := u.getInlineMessageID()
	if inlineMsgID == nil {
		return nil, fmt.Errorf("no inline message ID available for editing")
	}
	req := &tg.MessagesEditInlineBotMessageRequest{
		ID:          inlineMsgID,
		NoWebpage:   opts.NoWebpage,
		InvertMedia: opts.InvertMedia,
	}
	req.SetMedia(media)
	if caption != "" {
		req.SetMessage(caption)
	}
	if len(entities) > 0 {
		req.SetEntities(entities)
	}
	if opts.ReplyMarkup != nil {
		req.SetReplyMarkup(opts.ReplyMarkup)
	}
	if err := u.invokeInlineEdit(req); err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *Update) editInlineReplyMarkup(markup tg.ReplyMarkupClass) error {
	inlineMsgID := u.getInlineMessageID()
	if inlineMsgID == nil {
		return fmt.Errorf("no inline message ID available for editing")
	}
	req := &tg.MessagesEditInlineBotMessageRequest{
		ID: inlineMsgID,
	}
	if markup != nil {
		req.SetReplyMarkup(markup)
	}
	return u.invokeInlineEdit(req)
}

// Edit edits the current update's message text.
// Text can be a string or any type that can be formatted with %v.
// Uses the client's default parse mode from ClientOpts.ParseMode.
// For callback queries, edits the message that triggered the callback.
func (u *Update) Edit(text any, opts ...*EditOpts) (*types.Message, error) {
	if text == "" || text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	opt := functions.GetOptDef(&EditOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	var message string
	switch v := text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

	var parsedText string
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
			parsedText = result.Text
			entities = result.Entities
		} else {
			parsedText = message
		}
	} else {
		parsedText = message
	}

	if u.isInlineEdit() {
		return u.editInlineText(parsedText, entities, opt)
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	req := &tg.MessagesEditMessageRequest{
		ID:                   msgID,
		Message:              parsedText,
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
func (u *Update) EditMedia(media tg.InputMediaClass, opts ...*EditMediaOpts) (*types.Message, error) {
	if media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}

	opt := functions.GetOptDef(&EditMediaOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

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

	if u.isInlineEdit() {
		return u.editInlineMedia(media, caption, entities, opt)
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	req := &tg.MessagesEditMessageRequest{
		ID:                   msgID,
		Media:                media,
		Message:              caption,
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
func (u *Update) EditMediaWithFileID(fileID string, opts ...*EditMediaOpts) (*types.Message, error) {
	var caption string
	if len(opts) > 0 && opts[0] != nil {
		caption = opts[0].Caption
	}

	media, err := types.InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return u.EditMedia(media, opts...)
}

// EditCaption edits the caption of the current update's media message.
// Text can be a string or any type that can be formatted with %v.
// Uses the client's default parse mode from ClientOpts.ParseMode.
func (u *Update) EditCaption(text any, opts ...*EditOpts) (*types.Message, error) {
	if text == "" || text == nil {
		return nil, gotgErrors.ErrTextEmpty
	}

	opt := functions.GetOptDef(&EditOpts{}, opts...)

	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	var message string
	switch v := text.(type) {
	case string:
		message = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		message = fmt.Sprintf("%v", v)
	default:
		message = fmt.Sprintf("%v", v)
	}

	var parsedText string
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
			parsedText = result.Text
			entities = result.Entities
		} else {
			parsedText = message
		}
	} else {
		parsedText = message
	}

	if u.isInlineEdit() {
		return u.editInlineText(parsedText, entities, opt)
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	req := &tg.MessagesEditMessageRequest{
		ID:                   msgID,
		Message:              parsedText,
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
	return u.Ctx.EditMessage(u.ChatID(), req, connID)
}

// EditReplyMarkup edits only the reply markup of the current update's message.
// For callback queries, edits the message that triggered the callback.
func (u *Update) EditReplyMarkup(markup tg.ReplyMarkupClass) (*types.Message, error) {
	if u.isInlineEdit() {
		return nil, u.editInlineReplyMarkup(markup)
	}

	msgID, err := u.getEditMessageID()
	if err != nil {
		return nil, err
	}

	req := &tg.MessagesEditMessageRequest{
		ID:          msgID,
		ReplyMarkup: markup,
	}

	return u.Ctx.EditMessage(u.ChatID(), req, u.ConnectionID())
}

// EditInlineText edits an inline message's text.
// Works for inline callback queries and chosen inline results.
func (u *Update) EditInlineText(text any, opts ...*EditOpts) error {
	if text == "" || text == nil {
		return gotgErrors.ErrTextEmpty
	}

	opt := functions.GetOptDef(&EditOpts{}, opts...)
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	parsed, entities := parseTextEntities(fmt.Sprintf("%v", text), parseMode)
	_, err := u.editInlineText(parsed, entities, opt)
	return err
}

// EditInlineCaption edits an inline message's caption.
// Works for inline callback queries and chosen inline results.
func (u *Update) EditInlineCaption(caption any, opts ...*EditOpts) error {
	return u.EditInlineText(caption, opts...)
}

// EditInlineReplyMarkup edits only the reply markup of an inline message.
// Works for inline callback queries and chosen inline results.
func (u *Update) EditInlineReplyMarkup(markup tg.ReplyMarkupClass) error {
	return u.editInlineReplyMarkup(markup)
}

// EditInlineMedia edits the media of an inline message.
// Works for inline callback queries and chosen inline results.
func (u *Update) EditInlineMedia(media tg.InputMediaClass, opts ...*EditMediaOpts) error {
	if media == nil {
		return fmt.Errorf("media cannot be nil")
	}

	opt := functions.GetOptDef(&EditMediaOpts{}, opts...)
	parseMode := opt.ParseMode
	if parseMode == "" {
		parseMode = u.Ctx.DefaultParseMode
	}

	var caption string
	var entities []tg.MessageEntityClass
	if opt.Caption != "" {
		caption, entities = parseTextEntities(opt.Caption, parseMode)
	}

	_, err := u.editInlineMedia(media, caption, entities, opt)
	return err
}

func parseTextEntities(message string, parseMode string) (string, []tg.MessageEntityClass) {
	if parseMode == ModeNone {
		return message, nil
	}

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
		return result.Text, result.Entities
	}
	return message, nil
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
