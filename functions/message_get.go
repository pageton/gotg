package functions

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

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
	if peer == nil || peer.ID == 0 {
		// Try to resolve via API and retry
		if _, err := ResolveInputPeerByID(ctx, raw, p, chatID); err != nil {
			return nil, err
		}
		peer = p.GetPeerByID(chatID)
		if peer == nil || peer.ID == 0 {
			return nil, errors.ErrPeerNotFound
		}
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
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - peer: The channel peer to fetch messages from
//   - messageIds: List of message IDs to fetch
//
// Returns list of messages or an error.
func GetChannelMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, peer tg.InputChannelClass, messageIds []tg.InputMessageClass) (tg.MessageClassArray, error) {
	messages, err := raw.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
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
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - messageIds: List of message IDs to fetch
//
// Returns list of messages or an error.
func GetChatMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, messageIds []tg.InputMessageClass) (tg.MessageClassArray, error) {
	messages, err := raw.MessagesGetMessages(ctx, messageIds)
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
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
	}

	if _, isEmpty := inputPeer.(*tg.InputPeerEmpty); isEmpty {
		return nil, errors.ErrPeerNotFound
	}

	return raw.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer: inputPeer,
		ID:   messageID,
	})
}

// UnPinMessage unpins a specific message in a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID containing the message
//   - messageID: The message ID to unpin
//
// Returns an error if the operation fails.
func UnPinMessage(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, messageID int) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be empty")
	}
	if messageID == 0 {
		return fmt.Errorf("message ID cannot be empty")
	}

	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return err
		}
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

// UnPinAllMessages unpins all messages in a chat.
func UnPinAllMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be empty")
	}

	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return err
		}
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
	fromPeer, err := ResolveInputPeerByID(ctx, raw, p, fromChatID)
	if err != nil {
		return nil, err
	}

	toPeer, err := ResolveInputPeerByID(ctx, raw, p, toChatID)
	if err != nil {
		return nil, err
	}

	if request == nil {
		request = &tg.MessagesForwardMessagesRequest{}
	}

	if request.RandomID == nil {
		request.RandomID = make([]int64, len(request.ID))
		for i := 0; i < len(request.ID); i++ {
			request.RandomID[i] = GenerateRandomID()
		}
	}

	upd, err := raw.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
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
	if err != nil {
		return nil, err
	}
	switch u := upd.(type) {
	case *tg.Updates:
		SavePeersFromClassArray(p, u.Chats, u.Users)
	case *tg.UpdatesCombined:
		SavePeersFromClassArray(p, u.Chats, u.Users)
	}
	return upd, nil
}

// EditMessage edits a message in a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - peerStorage: Peer storage for resolving peer references
//   - chatID: The chat ID containing the message
//   - request: The edit message request parameters
//
// Returns the edited message or an error.
func EditMessage(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesEditMessageRequest, businessConnectionID ...string) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesEditMessageRequest{}
	}
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}
	var upds tg.UpdatesClass
	connID := getFirstString(businessConnectionID)
	if connID != "" {
		upds, err = invokeWithBusinessConnection(ctx, raw, connID, request)
	} else {
		upds, err = raw.MessagesEditMessage(ctx, request)
	}
	if err != nil {
		return nil, err
	}
	return ReturnEditMessageWithError(peerStorage, upds, nil)
}
