package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// GetChat returns full chat information for the provided chat ID.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID to get full information for
//
// Returns full chat information or an error.
func GetChat(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) (tg.ChatFullClass, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return nil, errors.ErrPeerNotFound
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
		return channel.FullChat, nil
	case *tg.InputPeerChat:
		chat, err := raw.MessagesGetFullChat(ctx, chatID)
		if err != nil {
			return nil, err
		}
		return chat.FullChat, nil
	default:
		return nil, errors.ErrNotChat
	}
}

// AddChatMembers adds multiple users to a chat.
//
// Parameters:
//   - context: Context for the API call
//   - client: The raw Telegram client
//   - chatPeer: The chat peer to add users to
//   - users: List of users to add
//   - forwardLimit: Maximum number of messages to forward from old chat
//
// Returns true if successful, or an error.
func AddChatMembers(context context.Context, client *tg.Client, chatPeer tg.InputPeerClass, users []tg.InputUserClass, forwardLimit int) (bool, error) {
	switch c := chatPeer.(type) {
	case *tg.InputPeerChat:
		for _, user := range users {
			user, ok := user.(*tg.InputUser)
			if ok {
				_, err := client.MessagesAddChatUser(context, &tg.MessagesAddChatUserRequest{
					ChatID: c.ChatID,
					UserID: &tg.InputUser{
						UserID:     user.UserID,
						AccessHash: user.AccessHash,
					},
					FwdLimit: forwardLimit,
				})
				if err != nil {
					return false, err
				}
			}
		}
		return true, nil
	case *tg.InputPeerChannel:
		_, err := client.ChannelsInviteToChannel(context, &tg.ChannelsInviteToChannelRequest{
			Channel: &tg.InputChannel{
				ChannelID:  c.ChannelID,
				AccessHash: c.AccessHash,
			},
			Users: users,
		})
		return err == nil, err
	}
	return false, nil
}

// ArchiveChats moves chats to the archive folder (folder ID 1).
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - peers: List of peers to archive
//
// Returns true if successful, or an error.
func ArchiveChats(ctx context.Context, client *tg.Client, peers []tg.InputPeerClass) (bool, error) {
	var folderPeers = make([]tg.InputFolderPeer, len(peers))
	for n, peer := range peers {
		folderPeers[n] = tg.InputFolderPeer{
			Peer:     peer,
			FolderID: 1,
		}
	}
	_, err := client.FoldersEditPeerFolders(ctx, folderPeers)
	return err == nil, err
}

// UnarchiveChats moves chats out of the archive folder (folder ID 0).
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - peers: List of peers to unarchive
//
// Returns true if successful, or an error.
func UnarchiveChats(ctx context.Context, client *tg.Client, peers []tg.InputPeerClass) (bool, error) {
	var folderPeers = make([]tg.InputFolderPeer, len(peers))
	for n, peer := range peers {
		folderPeers[n] = tg.InputFolderPeer{
			Peer:     peer,
			FolderID: 0,
		}
	}
	_, err := client.FoldersEditPeerFolders(ctx, folderPeers)
	return err == nil, err
}

// CreateChannel creates a new channel (supergroup or broadcast).
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
	return chats[0].(*tg.Channel), nil
}

// CreateChat creates a new group chat.
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
	return chats[0].(*tg.Chat), nil
}

// BanChatMember bans a user from a chat until the specified date.
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - chatPeer: The chat peer to ban user from
//   - userPeer: The user peer to ban
//   - untilDate: Unix timestamp until when the ban is active (0 for permanent)
//
// Returns updates confirming the action or an error.
func BanChatMember(ctx context.Context, client *tg.Client, chatPeer tg.InputPeerClass, userPeer *tg.InputPeerUser, untilDate int) (tg.UpdatesClass, error) {
	switch c := chatPeer.(type) {
	case *tg.InputPeerChannel:
		return client.ChannelsEditBanned(ctx, &tg.ChannelsEditBannedRequest{
			Channel: &tg.InputChannel{
				ChannelID:  c.ChannelID,
				AccessHash: c.AccessHash,
			},
			Participant: userPeer,
			BannedRights: tg.ChatBannedRights{
				UntilDate:    untilDate,
				ViewMessages: true,
				SendMessages: true,
				SendMedia:    true,
				SendStickers: true,
				SendGifs:     true,
				SendGames:    true,
				SendInline:   true,
				EmbedLinks:   true,
			},
		})
	case *tg.InputPeerChat:
		return client.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
			ChatID: c.ChatID,
			UserID: &tg.InputUser{
				UserID:     userPeer.UserID,
				AccessHash: userPeer.AccessHash,
			},
		})
	default:
		return &tg.Updates{}, nil
	}
}

// UnbanChatMember unbans a previously banned user from a channel.
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - chatPeer: The channel peer to unban user from
//   - userPeer: The user peer to unban
//
// Returns true if successful, or an error.
func UnbanChatMember(ctx context.Context, client *tg.Client, chatPeer *tg.InputPeerChannel, userPeer *tg.InputPeerUser) (bool, error) {
	_, err := client.ChannelsEditBanned(ctx, &tg.ChannelsEditBannedRequest{
		Channel: &tg.InputChannel{
			ChannelID:  chatPeer.ChannelID,
			AccessHash: chatPeer.AccessHash,
		},
		Participant: userPeer,
		BannedRights: tg.ChatBannedRights{
			UntilDate: 0,
		},
	})
	return err == nil, err
}

// PromoteChatMember promotes a user to admin in a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - chat: The chat peer storage
//   - user: The user peer storage to promote
//   - rights: Admin rights to grant
//   - title: Custom admin rank/title
//
// Returns true if successful, or an error.
func PromoteChatMember(ctx context.Context, client *tg.Client, chat, user *storage.Peer, rights tg.ChatAdminRights, title string) (bool, error) {
	rights.Other = true
	if chat.AccessHash != 0 {
		_, err := client.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel: &tg.InputChannel{
				ChannelID:  chat.GetID(),
				AccessHash: chat.AccessHash,
			},
			UserID: &tg.InputUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			},
			AdminRights: rights,
			Rank:        title,
		})
		return err == nil, err
	} else {
		_, err := client.MessagesEditChatAdmin(ctx, &tg.MessagesEditChatAdminRequest{
			ChatID: chat.GetID(),
			UserID: &tg.InputUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			},
			IsAdmin: true,
		})
		return err == nil, err
	}
}

// DemoteChatMember demotes an admin to regular member in a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - chat: The chat peer storage
//   - user: The user peer storage to demote
//   - rights: Admin rights to remove (set to empty)
//   - title: Custom admin rank/title to remove
//
// Returns true if successful, or an error.
func DemoteChatMember(ctx context.Context, client *tg.Client, chat, user *storage.Peer, rights tg.ChatAdminRights, title string) (bool, error) {
	rights.Other = false
	if chat.AccessHash != 0 {
		_, err := client.ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel: &tg.InputChannel{
				ChannelID:  chat.GetID(),
				AccessHash: chat.AccessHash,
			},
			UserID: &tg.InputUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			},
			AdminRights: rights,
			Rank:        title,
		})
		return err == nil, err
	} else {
		_, err := client.MessagesEditChatAdmin(ctx, &tg.MessagesEditChatAdminRequest{
			ChatID: chat.GetID(),
			UserID: &tg.InputUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			},
			IsAdmin: false,
		})
		return err == nil, err
	}
}
