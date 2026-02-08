package adapter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	gotgErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
)

type ChatMembersFilter = functions.ChatMembersFilter

const (
	FilterSearch   = functions.FilterSearch
	FilterRecent   = functions.FilterRecent
	FilterAdmins   = functions.FilterAdmins
	FilterBots     = functions.FilterBots
	FilterKicked   = functions.FilterKicked
	FilterBanned   = functions.FilterBanned
	FilterContacts = functions.FilterContacts
)

type GetChatMembersOpts = functions.GetChatMembersOpts

// GetChat returns tg.ChatClass of the provided chat id.
func (ctx *Context) GetChat(chatID int64) (tg.ChatClass, error) {
	return functions.GetChat(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID)
}

// GetFullChat returns full chat details for the provided chat ID.
func (ctx *Context) GetFullChat(chatID int64) (tg.ChatFullClass, error) {
	return functions.GetFullChat(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID)
}

// GetUser returns basic user information for the provided user ID.
func (ctx *Context) GetUser(userID int64) (*tg.User, error) {
	return functions.GetUser(ctx.Context, ctx.Raw, ctx.PeerStorage, userID)
}

// GetFullUser returns full user profile for the provided user ID.
func (ctx *Context) GetFullUser(userID int64) (*tg.UserFull, error) {
	return functions.GetFullUser(ctx.Context, ctx.Raw, ctx.PeerStorage, userID)
}

// GetChatMember fetches information about a chat member.
func (ctx *Context) GetChatMember(chatID, userID int64) (*types.Participant, error) {
	participant, err := functions.GetChatMember(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, userID)
	if err != nil {
		return nil, err
	}

	partUserID := functions.ExtractParticipantUserID(participant)
	if partUserID == 0 {
		return nil, fmt.Errorf("could not extract user ID from participant")
	}

	var tgUser *tg.User
	peer := ctx.PeerStorage.GetPeerByID(partUserID)
	if peer.Type == int(storage.TypeUser) {
		tgUser, err = functions.GetUser(ctx.Context, ctx.Raw, ctx.PeerStorage, partUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	} else {
		return nil, fmt.Errorf("user not found in peer storage")
	}

	return &types.Participant{
		User:        &types.User{User: *tgUser, Ctx: ctx.Context, RawClient: ctx.Raw, PeerStorage: ctx.PeerStorage, SelfID: ctx.Self.ID},
		Participant: participant,
		Status:      functions.ExtractParticipantStatus(participant),
		Rights:      functions.ExtractParticipantRights(participant),
		Title:       functions.ExtractParticipantTitle(participant),
		UserID:      partUserID,
		ChatID:      chatID,
	}, nil
}

func (ctx *Context) GetChatMembers(chatID int64, opts ...*functions.GetChatMembersOpts) ([]*types.Participant, error) {
	participants, err := functions.GetChatMembers(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, opts...)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Participant, 0, len(participants))
	for _, participant := range participants {
		partUserID := functions.ExtractParticipantUserID(participant)
		if partUserID == 0 {
			continue
		}

		var tgUser *tg.User
		peer := ctx.PeerStorage.GetPeerByID(partUserID)
		if peer.Type == int(storage.TypeUser) {
			tgUser, err = functions.GetUser(ctx.Context, ctx.Raw, ctx.PeerStorage, partUserID)
			if err != nil {
				continue
			}
		} else {
			continue
		}

		result = append(result, &types.Participant{
			User:        &types.User{User: *tgUser, Ctx: ctx.Context, RawClient: ctx.Raw, PeerStorage: ctx.PeerStorage, SelfID: ctx.Self.ID},
			Participant: participant,
			Status:      functions.ExtractParticipantStatus(participant),
			Rights:      functions.ExtractParticipantRights(participant),
			Title:       functions.ExtractParticipantTitle(participant),
			UserID:      partUserID,
			ChatID:      chatID,
		})
	}

	return result, nil
}

// GetMessages fetches messages from a chat by their IDs.
func (ctx *Context) GetMessages(chatID int64, messageIDs []tg.InputMessageClass) ([]tg.MessageClass, error) {
	return functions.GetMessages(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, messageIDs)
}

// BanChatMember bans a user from a chat until a specified date.
func (ctx *Context) BanChatMember(chatID, userID int64, untilDate int) (tg.UpdatesClass, error) {
	inputPeerChat, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return nil, err
	}
	switch inputPeerChat.(type) {
	case *tg.InputPeerChannel:
	case *tg.InputPeerChat:
	case *tg.InputPeerEmpty:
		return nil, gotgErrors.ErrPeerNotFound
	default:
		return nil, gotgErrors.ErrNotChat
	}
	var inputPeerUser *tg.InputPeerUser
	inputPeer, err := ctx.ResolveInputPeerByID(userID)
	if err != nil {
		return nil, err
	}
	switch p := inputPeer.(type) {
	case *tg.InputPeerUser:
		inputPeerUser = p
	case *tg.InputPeerEmpty:
		return nil, gotgErrors.ErrPeerNotFound
	default:
		return nil, gotgErrors.ErrNotUser
	}
	return functions.BanChatMember(ctx.Context, ctx.Raw, inputPeerChat, inputPeerUser, untilDate)
}

// UnbanChatMember unbans a previously banned user from a channel.
func (ctx *Context) UnbanChatMember(chatID, userID int64) (bool, error) {
	var inputPeerChat *tg.InputPeerChannel
	inputPeer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return false, err
	}
	switch p := inputPeer.(type) {
	case *tg.InputPeerChannel:
		inputPeerChat = p
	case *tg.InputPeerEmpty:
		return false, gotgErrors.ErrPeerNotFound
	default:
		return false, gotgErrors.ErrNotChannel
	}
	var inputPeerUser *tg.InputPeerUser
	inputPeer, err = ctx.ResolveInputPeerByID(userID)
	if err != nil {
		return false, err
	}
	switch p := inputPeer.(type) {

	case *tg.InputPeerUser:
		inputPeerUser = p
	case *tg.InputPeerEmpty:
		return false, gotgErrors.ErrPeerNotFound
	default:
		return false, gotgErrors.ErrNotUser
	}
	return functions.UnbanChatMember(ctx.Context, ctx.Raw, inputPeerChat, inputPeerUser)
}

// AddChatMembers adds multiple users to a chat.
func (ctx *Context) AddChatMembers(chatID int64, userIDs []int64, forwardLimit int) (bool, error) {
	inputPeerChat, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return false, err
	}
	switch inputPeerChat.(type) {
	case *tg.InputPeerChannel:
	case *tg.InputPeerChat:
	case *tg.InputPeerEmpty:
		return false, gotgErrors.ErrPeerNotFound
	default:
		return false, gotgErrors.ErrNotChat
	}
	userPeers := make([]tg.InputUserClass, len(userIDs))
	for i, uID := range userIDs {
		inputPeerUser, err := ctx.ResolveInputPeerByID(uID)
		if err != nil {
			return false, err
		}
		switch p := inputPeerUser.(type) {
		case *tg.InputPeerUser:
			userPeers[i] = &tg.InputUser{
				UserID:     p.UserID,
				AccessHash: p.AccessHash,
			}
		case *tg.InputPeerEmpty:
			return false, gotgErrors.ErrPeerNotFound
		default:
			return false, gotgErrors.ErrNotUser
		}
	}
	return functions.AddChatMembers(ctx.Context, ctx.Raw, inputPeerChat, userPeers, forwardLimit)
}

// ArchiveChats moves chats to the archive folder (folder ID 1).
func (ctx *Context) ArchiveChats(chatIDs []int64) (bool, error) {
	chatPeers := make([]tg.InputPeerClass, len(chatIDs))
	for i, chatID := range chatIDs {
		inputPeer, err := ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return false, err
		}
		switch inputPeer.(type) {
		case *tg.InputPeerChannel:
		case *tg.InputPeerChat:
		case *tg.InputPeerUser:
		case *tg.InputPeerEmpty:
			return false, gotgErrors.ErrPeerNotFound
		default:
			return false, gotgErrors.ErrNotChat
		}
		chatPeers[i] = inputPeer
	}
	return functions.ArchiveChats(ctx.Context, ctx.Raw, chatPeers)
}

// UnarchiveChats moves chats out of the archive folder (folder ID 0).
func (ctx *Context) UnarchiveChats(chatIDs []int64) (bool, error) {
	chatPeers := make([]tg.InputPeerClass, len(chatIDs))
	for i, chatID := range chatIDs {
		inputPeer, err := ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return false, err
		}
		switch inputPeer.(type) {
		case *tg.InputPeerChannel:
		case *tg.InputPeerChat:
		case *tg.InputPeerUser:
		case *tg.InputPeerEmpty:
			return false, gotgErrors.ErrPeerNotFound
		default:
			return false, gotgErrors.ErrNotChat
		}
		chatPeers[i] = inputPeer
	}
	return functions.UnarchiveChats(ctx.Context, ctx.Raw, chatPeers)
}

// CreateChannel creates a new channel (supergroup or broadcast).
func (ctx *Context) CreateChannel(title, about string, broadcast bool) (*tg.Channel, error) {
	return functions.CreateChannel(ctx.Context, ctx.Raw, ctx.PeerStorage, title, about, broadcast)
}

// CreateChat creates a new group chat.
func (ctx *Context) CreateChat(title string, userIDs []int64) (*tg.Chat, error) {
	userPeers := make([]tg.InputUserClass, len(userIDs))
	for i, uID := range userIDs {
		userPeer := ctx.ResolvePeerByID(uID)
		if userPeer.ID == 0 {
			return nil, gotgErrors.ErrPeerNotFound
		}
		if userPeer.Type != int(storage.TypeUser) {
			return nil, gotgErrors.ErrNotUser
		}
		userPeers[i] = &tg.InputUser{
			UserID:     userPeer.ID,
			AccessHash: userPeer.AccessHash,
		}
	}
	return functions.CreateChat(ctx.Context, ctx.Raw, ctx.PeerStorage, title, userPeers)
}

// DeleteMessages deletes messages in a chat.
func (ctx *Context) DeleteMessages(chatID int64, messageIDs []int) error {
	return functions.DeleteMessages(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, messageIDs)
}

// ForwardMessage forwards messages from one chat to another.
// Deprecated: use ForwardMessages instead.
func (ctx *Context) ForwardMessage(fromChatID, toChatID int64, request *tg.MessagesForwardMessagesRequest) (tg.UpdatesClass, error) {
	return ctx.ForwardMessages(fromChatID, toChatID, request)
}

// ForwardMessages forwards messages from one chat to another.
func (ctx *Context) ForwardMessages(fromChatID, toChatID int64, request *tg.MessagesForwardMessagesRequest) (tg.UpdatesClass, error) {
	return functions.ForwardMessages(ctx.Context, ctx.Raw, ctx.PeerStorage, fromChatID, toChatID, request)
}

// EditAdminOpts contains options for editing admin rights.
type EditAdminOpts struct {
	AdminRights tg.ChatAdminRights
	AdminTitle  string
}

// PromoteChatMember promotes a user to admin in a chat.
func (ctx *Context) PromoteChatMember(chatID, userID int64, opts *EditAdminOpts) (bool, error) {
	peerChat := ctx.ResolvePeerByID(chatID)
	if peerChat.ID == 0 {
		return false, fmt.Errorf("chat: %w", gotgErrors.ErrPeerNotFound)
	}
	peerUser := ctx.ResolvePeerByID(userID)
	if peerUser.ID == 0 {
		return false, fmt.Errorf("user: %w", gotgErrors.ErrPeerNotFound)
	}
	if opts == nil {
		opts = &EditAdminOpts{}
	}
	return functions.PromoteChatMember(ctx.Context, ctx.Raw, peerChat, peerUser, opts.AdminRights, opts.AdminTitle)
}

// DemoteChatMember demotes an admin to regular member in a chat.
func (ctx *Context) DemoteChatMember(chatID, userID int64, opts *EditAdminOpts) (bool, error) {
	peerChat := ctx.ResolvePeerByID(chatID)
	if peerChat.ID == 0 {
		return false, fmt.Errorf("chat: %w", gotgErrors.ErrPeerNotFound)
	}
	peerUser := ctx.ResolvePeerByID(userID)
	if peerUser.ID == 0 {
		return false, fmt.Errorf("user: %w", gotgErrors.ErrPeerNotFound)
	}
	if opts == nil {
		opts = &EditAdminOpts{}
	}
	return functions.DemoteChatMember(ctx.Context, ctx.Raw, peerChat, peerUser, opts.AdminRights, opts.AdminTitle)
}

// ResolveUsername resolves a @username to get peer information.
func (ctx *Context) ResolveUsername(username string) (types.EffectiveChat, error) {
	return ctx.extractContactResolvedPeer(
		ctx.Raw.ContactsResolveUsername(
			ctx,
			&tg.ContactsResolveUsernameRequest{
				Username: strings.TrimPrefix(
					username,
					"@",
				),
			},
		),
	)
}

func (ctx *Context) extractContactResolvedPeer(p *tg.ContactsResolvedPeer, err error) (types.EffectiveChat, error) {
	if err != nil {
		return &types.EmptyUC{}, err
	}
	functions.SavePeersFromClassArray(ctx.PeerStorage, p.Chats, p.Users)
	switch p.Peer.(type) {
	case *tg.PeerChannel:
		if len(p.Chats) == 0 {
			return &types.EmptyUC{}, errors.New("peer info not found in the resolved Chats")
		}
		switch chat := p.Chats[0].(type) {
		case *tg.Channel:
			c := &types.Channel{
				Channel:     *chat,
				Ctx:         ctx.Context,
				RawClient:   ctx.Raw,
				PeerStorage: ctx.PeerStorage,
				SelfID:      ctx.Self.ID,
			}
			return c, nil
		case *tg.ChannelForbidden:
			return &types.EmptyUC{}, errors.New("peer could not be resolved because Channel Forbidden")
		}
	case *tg.PeerUser:
		if len(p.Users) == 0 {
			return &types.EmptyUC{}, errors.New("peer info not found in the resolved Chats")
		}
		switch user := p.Users[0].(type) {
		case *tg.User:
			c := &types.User{
				User:        *user,
				Ctx:         ctx.Context,
				RawClient:   ctx.Raw,
				PeerStorage: ctx.PeerStorage,
				SelfID:      ctx.Self.ID,
			}
			return c, nil
		}
	}
	return &types.EmptyUC{}, errors.New("contact not found")
}

// GetUserProfilePhotos retrieves profile photos for a user.
func (ctx *Context) GetUserProfilePhotos(userID int64, opts *tg.PhotosGetUserPhotosRequest) ([]tg.PhotoClass, error) {
	return functions.GetUserProfilePhotos(ctx.Context, ctx.Raw, ctx.PeerStorage, userID, opts)
}
