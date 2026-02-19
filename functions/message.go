package functions

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

func invokeWithBusinessConnection(ctx context.Context, raw *tg.Client, connectionID string, request bin.Object) (tg.UpdatesClass, error) {
	var result tg.UpdatesBox
	err := raw.Invoker().Invoke(ctx, &tg.InvokeWithBusinessConnectionRequest{
		ConnectionID: connectionID,
		Query:        request,
	}, &result)
	if err != nil {
		return nil, err
	}
	return result.Updates, nil
}

func getFirstString(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

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
func SendMessage(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendMessageRequest, businessConnectionID ...string) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMessageRequest{}
	}
	request.RandomID = GenerateRandomID()
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}
	m := &tg.Message{}
	m.Message = request.Message
	var u tg.UpdatesClass
	connID := getFirstString(businessConnectionID)
	if connID != "" {
		u, err = invokeWithBusinessConnection(ctx, raw, connID, request)
	} else {
		u, err = raw.MessagesSendMessage(ctx, request)
	}
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
func SendMedia(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendMediaRequest, businessConnectionID ...string) (*tg.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMediaRequest{}
	}
	request.RandomID = GenerateRandomID()
	if request.Peer == nil {
		request.Peer, err = ResolveInputPeerByID(ctx, raw, peerStorage, chatID)
		if err != nil {
			return nil, err
		}
	}
	m := &tg.Message{}
	m.Message = request.Message
	var u tg.UpdatesClass
	connID := getFirstString(businessConnectionID)
	if connID != "" {
		u, err = invokeWithBusinessConnection(ctx, raw, connID, request)
	} else {
		u, err = raw.MessagesSendMedia(ctx, request)
	}
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
	m := &tg.Message{}
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
func SendMultiMedia(ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, chatID int64, request *tg.MessagesSendMultiMediaRequest, businessConnectionID ...string) (*tg.Message, error) {
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
	var u tg.UpdatesClass
	connID := getFirstString(businessConnectionID)
	if connID != "" {
		u, err = invokeWithBusinessConnection(ctx, raw, connID, request)
	} else {
		u, err = raw.MessagesSendMultiMedia(ctx, request)
	}
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
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return err
		}
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
