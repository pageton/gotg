package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

func GetMessages(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, mids []tg.InputMessageClass) (tg.MessageClassArray, error) {
	peer := p.GetPeerByID(chatID)
	if peer.ID == 0 {
		return nil, errors.ErrPeerNotFound
	}
	switch storage.EntityType(peer.Type) {
	case storage.TypeChannel:
		return GetChannelMessages(ctx, raw, p, &tg.InputChannel{
			ChannelID:  peer.GetID(),
			AccessHash: peer.AccessHash,
		}, mids)
	default:
		return GetChatMessages(ctx, raw, p, mids)
	}
}

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
	case *tg.MessagesChannelMessages:
		SavePeersFromClassArray(p, m.Chats, m.Users)
		return m.Messages, nil
	default:
		return nil, nil
	}
}
