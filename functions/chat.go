package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// GetFullChat returns full chat details for the provided chat ID.
// Uses ChannelsGetFullChannel for channels or MessagesGetFullChat for groups.
//
// Example:
//
//	fullChat, err := functions.GetFullChat(ctx, client.Raw, client.PeerStorage, chatID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	switch fc := fullChat.(type) {
//	case *tg.ChannelFull:
//	    fmt.Printf("About: %s\n", fc.About)
//	case *tg.ChatFull:
//	    fmt.Printf("Participants: %d\n", len(fc.Participants.(*tg.ChatParticipants).Participants))
//	}
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID to get full information for
//
// Returns full chat information or an error.
func GetFullChat(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) (tg.ChatFullClass, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
	}
	switch peer := inputPeer.(type) {
	case *tg.InputPeerChannel:
		channel, err := raw.ChannelsGetFullChannel(ctx, &tg.InputChannel{
			ChannelID:  peer.ChannelID,
			AccessHash: peer.AccessHash,
		})
		if err != nil {
			return nil, err
		}
		SavePeersFromClassArray(p, channel.Chats, channel.Users)
		return channel.FullChat, nil
	case *tg.InputPeerChat:
		chat, err := raw.MessagesGetFullChat(ctx, chatID)
		if err != nil {
			return nil, err
		}
		SavePeersFromClassArray(p, chat.Chats, chat.Users)
		return chat.FullChat, nil
	default:
		return nil, errors.ErrNotChat
	}
}

// GetChat returns basic chat information for the provided chat ID.
// Uses MessagesGetChats for efficient single-chat lookup.
//
// Example:
//
//	chat, err := functions.GetChat(ctx, client.Raw, chatID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Title: %s, Participants: %d\n",
//	    chat.Title(), len(chat.ParticipantsCount))
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - chatID: The chat ID to get basic information for
//
// Returns basic Chat or Channel information (no full details).
func GetChat(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) (tg.ChatClass, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
	}
	switch peer := inputPeer.(type) {
	case *tg.InputPeerChannel:
		chatsClass, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			},
		})
		if err != nil {
			return nil, err
		}
		SavePeersFromClassArray(p, chatsClass.MapChats(), nil)
		chat, ok := chatsClass.MapChats().First()
		if !ok {
			return nil, errors.ErrPeerNotFound
		}
		return chat, nil
	case *tg.InputPeerChat:
		chatsClass, err := raw.MessagesGetChats(ctx, []int64{chatID})
		if err != nil {
			return nil, err
		}
		SavePeersFromClassArray(p, chatsClass.MapChats(), nil)
		chat, ok := chatsClass.MapChats().First()
		if !ok {
			return nil, errors.ErrPeerNotFound
		}
		return chat, nil
	default:
		return nil, errors.ErrNotChat
	}
}

// CreateChannel creates a new channel (supergroup or broadcast).
//
// Example:
//
//	channel, err := functions.CreateChannel(ctx, client.Raw, client.PeerStorage, "My Channel", "Channel description", false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created channel ID: %d\n", channel.ID)
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - title: The channel title
//   - about: The channel description
//   - broadcast: Whether this is a broadcast channel (false for supergroup)
//
// Returns the created channel or an error.
func CreateChannel(ctx context.Context, client *tg.Client, p *storage.PeerStorage, title, about string, broadcast bool) (*tg.Channel, error) {
	udps, err := client.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
		Title:     title,
		About:     about,
		Broadcast: broadcast,
	})
	if err != nil {
		return nil, err
	}
	_, chats, _ := getUpdateFromUpdates(udps, p)
	if len(chats) == 0 {
		return nil, errors.ErrPeerNotFound
	}
	channel, ok := chats[0].(*tg.Channel)
	if !ok {
		return nil, errors.ErrNotChat
	}
	return channel, nil
}

// CreateChat creates a new group chat.
//
// Example:
//
//	users := []tg.InputUserClass{
//	    &tg.InputUser{UserID: 12345678, AccessHash: 1234567890},
//	}
//	chat, err := functions.CreateChat(ctx, client.Raw, client.PeerStorage, "My Group", users)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created chat ID: %d\n", chat.ID)
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - title: The chat title
//   - users: List of users to add to the chat
//
// Returns the created chat or an error.
func CreateChat(ctx context.Context, client *tg.Client, p *storage.PeerStorage, title string, users []tg.InputUserClass) (*tg.Chat, error) {
	udps, err := client.MessagesCreateChat(ctx, &tg.MessagesCreateChatRequest{
		Users: users,
		Title: title,
	})
	if err != nil {
		return nil, err
	}
	_, chats, _ := getUpdateFromUpdates(udps.Updates, p)
	if len(chats) == 0 {
		return nil, errors.ErrPeerNotFound
	}
	chat, ok := chats[0].(*tg.Chat)
	if !ok {
		return nil, errors.ErrNotChat
	}
	return chat, nil
}
