package functions

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// SendMessage sends a text message to a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peerStorage: Peer storage for resolving peer references
//   - chatID: The chat ID to send message to
//   - request: The message send request parameters
//
// Returns the sent message or an error.
func SendMessage(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendMessageRequest) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMessageRequest{}
	}
	request.RandomID = rand.Int63()
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}
	var m = &tg.Message{}
	m.Message = request.Message
	u, err := raw.MessagesSendMessage(ctx, request)
	message, err := ReturnNewMessageWithError(m, u, peerStorage, err)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// SendMedia sends a media message to a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peerStorage: Peer storage for resolving peer references
//   - chatID: The chat ID to send media to
//   - request: The media send request parameters
//
// Returns the sent message or an error.
func SendMedia(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendMediaRequest) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMediaRequest{}
	}
	request.RandomID = rand.Int63()
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}

	var m = &tg.Message{}
	m.Message = request.Message
	u, err := raw.MessagesSendMedia(ctx, request)
	message, err := ReturnNewMessageWithError(m, u, peerStorage, err)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// SendReaction sends a reaction to a message.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peerStorage: Peer storage for resolving peer references
//   - chatID: The chat ID containing the message
//   - request: The reaction send request parameters
//
// Returns the updated message or an error.
func SendReaction(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendReactionRequest) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendReactionRequest{}
	}
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}
	var m = &tg.Message{}
	u, err := raw.MessagesSendReaction(ctx, request)
	message, err := ReturnNewMessageWithError(m, u, peerStorage, err)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// SendMultiMedia sends multiple media items (album) to a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peerStorage: Peer storage for resolving peer references
//   - chatID: The chat ID to send album to
//   - request: The multi-media send request parameters
//
// Returns the sent message or an error.
func SendMultiMedia(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendMultiMediaRequest) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMultiMediaRequest{}
	}
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}
	u, err := raw.MessagesSendMultiMedia(ctx, request)
	message, err := ReturnNewMessageWithError(&tg.Message{}, u, peerStorage, err)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// DeleteMessages deletes messages in a chat with given chat ID and message IDs.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID containing the messages
//   - messageIDs: List of message IDs to delete
//
// Returns an error if the operation fails.
func DeleteMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, messageIDs []int) error {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return errors.ErrPeerNotFound
	}
	switch peer := inputPeer.(type) {
	case *tg.InputPeerChannel:
		_, err := raw.ChannelsDeleteMessages(ctx, &tg.ChannelsDeleteMessagesRequest{
			Channel: &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			},
			ID: messageIDs,
		})
		return err
	case *tg.InputPeerChat, *tg.InputPeerUser:
		_, err := raw.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
			Revoke: true,
			ID:     messageIDs,
		})
		return err
	default:
		return errors.ErrNotChat
	}
}

// GetMessages fetches messages from a chat by their IDs.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID to fetch messages from
//   - messageIDs: List of message IDs to fetch
//
// Returns list of messages or an error.
func GetMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, messageIDs []tg.InputMessageClass) (tg.MessageClassArray, error) {
	peer := p.GetPeerByID(chatID)
	if peer.ID == 0 {
		return nil, errors.ErrPeerNotFound
	}
	switch storage.EntityType(peer.Type) {
	case storage.TypeChannel:
		return GetChannelMessages(ctx, raw, p, &tg.InputChannel{
			ChannelID:  peer.GetID(),
			AccessHash: peer.AccessHash,
		}, messageIDs)
	default:
		return GetChatMessages(ctx, raw, p, messageIDs)
	}
}

// GetChannelMessages fetches messages from a channel.
//
// Parameters:
//   - context: Context for the API call
//   - client: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - peer: The channel peer to fetch messages from
//   - messageIds: List of message IDs to fetch
//
// Returns list of messages or an error.
func GetChannelMessages(context context.Context, client *tg.Client, p *storage.PeerStorage, peer tg.InputChannelClass, messageIds []tg.InputMessageClass) (tg.MessageClassArray, error) {
	messages, err := client.ChannelsGetMessages(context, &tg.ChannelsGetMessagesRequest{
		Channel: peer,
		ID:      messageIds,
	})
	if err != nil {
		return nil, err
	}
	switch m := messages.(type) {
	case *tg.MessagesMessages:
		SavePeersFromClassArray(p, m.Chats, m.Users)
		return m.Messages, nil
	case *tg.MessagesMessagesSlice:
		SavePeersFromClassArray(p, m.Chats, m.Users)
		return m.Messages, nil
	case *tg.MessagesChannelMessages:
		SavePeersFromClassArray(p, m.Chats, m.Users)
		return m.Messages, nil
	default:
		return nil, nil
	}
}

// GetChatMessages fetches messages from a regular chat.
//
// Parameters:
//   - context: Context for the API call
//   - client: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - messageIds: List of message IDs to fetch
//
// Returns list of messages or an error.
func GetChatMessages(context context.Context, client *tg.Client, p *storage.PeerStorage, messageIds []tg.InputMessageClass) (tg.MessageClassArray, error) {
	messages, err := client.MessagesGetMessages(context, messageIds)
	if err != nil {
		return nil, err
	}
	switch m := messages.(type) {
	case *tg.MessagesMessages:
		SavePeersFromClassArray(p, m.Chats, m.Users)
		return m.Messages, nil
	case *tg.MessagesMessagesSlice:
		SavePeersFromClassArray(p, m.Chats, m.Users)
		return m.Messages, nil
	default:
		return nil, nil
	}
}

// PinMessage pins a message in a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID containing the message
//   - messageID: The message ID to pin
//
// Returns updates confirming the action or an error.
func PinMessage(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, messageID int) (tg.UpdatesClass, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat ID cannot be empty")
	}
	if messageID == 0 {
		return nil, fmt.Errorf("message ID cannot be empty")
	}

	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return nil, errors.ErrPeerNotFound
	}

	if _, isEmpty := inputPeer.(*tg.InputPeerEmpty); isEmpty {
		return nil, errors.ErrPeerNotFound
	}

	return raw.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer: inputPeer,
		ID:   messageID,
	})
}

// UnpinMessage unpins a specific message in a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID containing the message
//   - messageID: The message ID to unpin
//
// Returns an error if the operation fails.
func UnpinMessage(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, messageID int) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be empty")
	}
	if messageID == 0 {
		return fmt.Errorf("message ID cannot be empty")
	}

	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return errors.ErrPeerNotFound
	}

	if _, isEmpty := inputPeer.(*tg.InputPeerEmpty); isEmpty {
		return errors.ErrPeerNotFound
	}

	_, err := raw.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:  inputPeer,
		ID:    messageID,
		Unpin: true,
	})
	return err
}

// UnpinAllMessages unpins all messages in a chat.
func UnpinAllMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be empty")
	}

	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return errors.ErrPeerNotFound
	}

	if _, isEmpty := inputPeer.(*tg.InputPeerEmpty); isEmpty {
		return errors.ErrPeerNotFound
	}

	_, err := raw.MessagesUnpinAllMessages(ctx, &tg.MessagesUnpinAllMessagesRequest{
		Peer: inputPeer,
	})
	return err
}

// ForwardMessages forwards messages from one chat to another.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - fromChatID: The source chat ID to forward from
//   - toChatID: The destination chat ID to forward to
//   - request: The forward messages request parameters
//
// Returns updates confirming the action or an error.
func ForwardMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, fromChatID, toChatID int64, request *tg.MessagesForwardMessagesRequest) (tg.UpdatesClass, error) {
	fromPeer := GetInputPeerClassFromID(p, fromChatID)
	if fromPeer == nil {
		return nil, errors.ErrPeerNotFound
	}

	toPeer := GetInputPeerClassFromID(p, toChatID)
	if toPeer == nil {
		return nil, errors.ErrPeerNotFound
	}

	if request == nil {
		request = &tg.MessagesForwardMessagesRequest{}
	}

	if request.RandomID == nil {
		request.RandomID = make([]int64, len(request.ID))
		for i := 0; i < len(request.ID); i++ {
			request.RandomID[i] = rand.Int63()
		}
	}

	return raw.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
		RandomID:           request.RandomID,
		ID:                 request.ID,
		FromPeer:           fromPeer,
		ToPeer:             toPeer,
		DropAuthor:         request.DropAuthor,
		Silent:             request.Silent,
		Background:         request.Background,
		WithMyScore:        request.WithMyScore,
		DropMediaCaptions:  request.DropMediaCaptions,
		Noforwards:         request.Noforwards,
		TopMsgID:           request.TopMsgID,
		ScheduleDate:       request.ScheduleDate,
		SendAs:             request.SendAs,
		QuickReplyShortcut: request.QuickReplyShortcut,
	})
}
