package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/types"
)

// EditMessage edits a previously sent message.
//
// Use this method to edit message text, media, caption, entities, or reply markup.
// Not all message types can be edited - see Telegram's API documentation for details.
//
// Parameters:
//   - chatID: The chat ID containing the message
//   - request: Telegram's MessagesEditMessageRequest with edits to apply
//
// Returns the edited Message or an error.
func (ctx *Context) EditMessage(chatID int64, request *tg.MessagesEditMessageRequest, businessConnectionID ...string) (*types.Message, error) {
	message, err := functions.EditMessage(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, request, businessConnectionID...)
	if err != nil {
		ctx.emitOutgoing(ActionEdit, StatusFailed, nil, request.ID, chatID, request.Peer, err)
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	ctx.emitOutgoing(ActionEdit, StatusSucceeded, msg, message.ID, chatID, request.Peer, nil)
	return msg, nil
}

// EditCaption edits the caption of a media message.
//
// Parameters:
//   - chatID: The chat ID containing the message
//   - messageID: The ID of the message to edit
//   - caption: The new caption text
//   - entities: Optional formatting entities for the caption
//
// Returns the edited Message or an error.
func (ctx *Context) EditCaption(chatID int64, messageID int, caption string, entities []tg.MessageEntityClass) (*types.Message, error) {
	req := &tg.MessagesEditMessageRequest{
		ID:       messageID,
		Message:  caption,
		Entities: entities,
	}
	return ctx.EditMessage(chatID, req)
}

// EditReplyMarkup edits only the reply markup (inline keyboard) of a message.
//
// Parameters:
//   - chatID: The chat ID containing the message
//   - messageID: The ID of the message to edit
//   - markup: The new reply markup (inline keyboard or force reply)
//
// Returns the edited Message or an error.
func (ctx *Context) EditReplyMarkup(chatID int64, messageID int, markup tg.ReplyMarkupClass) (*types.Message, error) {
	req := &tg.MessagesEditMessageRequest{
		ID:          messageID,
		ReplyMarkup: markup,
	}
	return ctx.EditMessage(chatID, req)
}
