package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// AddChatMembers adds multiple users to a chat.
//
// Example:
//
//	users := []tg.InputUserClass{
//	    &tg.InputUser{UserID: 12345678, AccessHash: 1234567890},
//	    &tg.InputUser{UserID: 87654321, AccessHash: 9876543210},
//	}
//	success, err := functions.AddChatMembers(ctx, client.Raw, chatPeer, users, 50)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - chatPeer: The chat peer to add users to
//   - users: List of users to add
//   - forwardLimit: Maximum number of messages to forward from old chat
//
// Returns true if successful, or an error.
func AddChatMembers(ctx context.Context, raw *tg.Client, chatPeer tg.InputPeerClass, users []tg.InputUserClass, forwardLimit int) (bool, error) {
	switch c := chatPeer.(type) {
	case *tg.InputPeerChat:
		for _, user := range users {
			inputUser, ok := user.(*tg.InputUser)
			if ok {
				_, err := raw.MessagesAddChatUser(ctx, &tg.MessagesAddChatUserRequest{
					ChatID: c.ChatID,
					UserID: &tg.InputUser{
						UserID:     inputUser.UserID,
						AccessHash: inputUser.AccessHash,
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
		_, err := raw.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
			Channel: &tg.InputChannel{
				ChannelID:  c.ChannelID,
				AccessHash: c.AccessHash,
			},
			Users: users,
		})
		return err == nil, err
	default:
		return false, nil
	}
}

// LeaveChannel leaves the current user from a channel or chat.
//
// Example:
//
//	success, err := functions.LeaveChannel(ctx, client.Raw, chatPeer)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - ctx: Context for API call
//   - client: The raw Telegram client
//   - chatPeer: The channel/chat peer to leave from
//
// Returns true if successful, or an error.
func LeaveChannel(ctx context.Context, client *tg.Client, chatPeer tg.InputPeerClass) (tg.UpdatesClass, error) {
	switch c := chatPeer.(type) {
	case *tg.InputPeerChannel:
		return client.ChannelsLeaveChannel(ctx, &tg.InputChannel{
			ChannelID:  c.ChannelID,
			AccessHash: c.AccessHash,
		})
	case *tg.InputPeerChat:
		return client.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
			ChatID: c.ChatID,
		})
	default:
		return &tg.Updates{}, nil
	}
}

// BanChatMember bans a user from a chat until the specified date.
//
// Example:
//
//	userPeer := &tg.InputPeerUser{UserID: 12345678, AccessHash: 1234567890}
//	updates, err := functions.BanChatMember(ctx, client.Raw, chatPeer, userPeer, 0)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
// Example:
//
//	channelPeer := &tg.InputPeerChannel{ChannelID: 12345678, AccessHash: 1234567890}
//	userPeer := &tg.InputPeerUser{UserID: 87654321, AccessHash: 9876543210}
//	success, err := functions.UnbanChatMember(ctx, client.Raw, channelPeer, userPeer)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
// Example:
//
//	rights := tg.ChatAdminRights{
//	    ChangeInfo:     true,
//	    BanUsers:       true,
//	    DeleteMessages: true,
//	    PinMessages:    true,
//	}
//	success, err := functions.PromoteChatMember(ctx, client.Raw, chatPeer, userPeer, rights, "Moderator")
//	if err != nil {
//	    log.Fatal(err)
//	}
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
// Example:
//
//	rights := tg.ChatAdminRights{} // Empty rights to remove all admin privileges
//	success, err := functions.DemoteChatMember(ctx, client.Raw, chatPeer, userPeer, rights, "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - ctx: Context for API call
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
