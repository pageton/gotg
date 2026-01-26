package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/types"
)

// SendMessage sends a text message to a chat.
//
// This is the primary method for sending messages. It supports text messages,
// formatting entities, inline keyboards, and more through the request parameter.
//
// Parameters:
//   - chatID: The chat ID to send the message to
//   - request: Telegram's MessagesSendMessageRequest containing message content
//
// Returns the sent Message or an error.
func (ctx *Context) SendMessage(chatID int64, request *tg.MessagesSendMessageRequest) (*types.Message, error) {
	message, err := functions.SendMessage(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, request)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// SendMedia sends media (photo, video, document, etc.) to a chat.
//
// Use this method to send any type of media including photos, videos,
// documents, audio files, and other media types supported by Telegram.
//
// Parameters:
//   - chatID: The chat ID to send the media to
//   - request: Telegram's MessagesSendMediaRequest containing media and metadata
//
// Returns the sent Message or an error.
func (ctx *Context) SendMedia(chatID int64, request *tg.MessagesSendMediaRequest) (*types.Message, error) {
	message, err := functions.SendMedia(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, request)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// SendReaction sends or removes emoji reactions to a message.
//
// Reactions allow users to quickly respond to messages with emoji.
// Set reaction to an empty slice to remove all reactions.
//
// Parameters:
//   - chatID: The chat ID containing the message
//   - request: Telegram's MessagesSendReactionRequest containing reactions
//
// Returns the Message with updated reactions or an error.
func (ctx *Context) SendReaction(chatID int64, request *tg.MessagesSendReactionRequest) (*types.Message, error) {
	message, err := functions.SendReaction(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, request)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// SendMultiMedia sends multiple media items as an album or grouped media.
//
// Use this method to send 2-10 media items as an album. Media items will
// be grouped together in the chat and displayed as a swipeable album.
//
// Parameters:
//   - chatID: The chat ID to send the album to
//   - request: Telegram's MessagesSendMultiMediaRequest containing media items
//
// Returns the sent Message album or an error.
func (ctx *Context) SendMultiMedia(chatID int64, request *tg.MessagesSendMultiMediaRequest) (*types.Message, error) {
	message, err := functions.SendMultiMedia(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, request)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}
