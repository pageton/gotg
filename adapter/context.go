package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	mtp_errors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
)

// generateRandomID generates a random int64 for use in Telegram API calls.
// Random IDs are required by Telegram for duplicate request prevention.
func (ctx *Context) generateRandomID() int64 {
	return ctx.random.Int63()
}

// SendMessage invokes method messages.sendMessage#d9d75a4 returning error if any.
func (ctx *Context) SendMessage(chatID int64, request *tg.MessagesSendMessageRequest) (*types.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMessageRequest{}
	}
	request.RandomID = ctx.generateRandomID()
	if request.Peer == nil {
		request.Peer, err = ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return nil, err
		}
	}
	var m = &tg.Message{}
	m.Message = request.Message
	u, err := ctx.Raw.MessagesSendMessage(ctx, request)
	message, err := functions.ReturnNewMessageWithError(m, u, ctx.PeerStorage, err)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// SendMedia invokes method messages.sendMedia#e25ff8e0 returning error if any. Send a media
func (ctx *Context) SendMedia(chatID int64, request *tg.MessagesSendMediaRequest) (*types.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMediaRequest{}
	}
	request.RandomID = ctx.generateRandomID()
	if request.Peer == nil {
		request.Peer, err = ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return nil, err
		}
	}

	var m = &tg.Message{}
	m.Message = request.Message
	u, err := ctx.Raw.MessagesSendMedia(ctx, request)
	message, err := functions.ReturnNewMessageWithError(m, u, ctx.PeerStorage, err)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// SendReaction invokes method messages.sendReaction#25690ce4 returning error if any.
func (ctx *Context) SendReaction(chatID int64, request *tg.MessagesSendReactionRequest) (*types.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendReactionRequest{}
	}
	if request.Peer == nil {
		request.Peer, err = ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return nil, err
		}
	}
	var m = &tg.Message{}
	// m.Message = request.Reaction
	u, err := ctx.Raw.MessagesSendReaction(ctx, request)
	message, err := functions.ReturnNewMessageWithError(m, u, ctx.PeerStorage, err)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// SendMultiMedia invokes method messages.sendMultiMedia#f803138f returning error if any. Send an album or grouped media¹
func (ctx *Context) SendMultiMedia(chatID int64, request *tg.MessagesSendMultiMediaRequest) (*types.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesSendMultiMediaRequest{}
	}
	if request.Peer == nil {
		request.Peer, err = ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return nil, err
		}
	}
	u, err := ctx.Raw.MessagesSendMultiMedia(ctx, request)
	message, err := functions.ReturnNewMessageWithError(&tg.Message{}, u, ctx.PeerStorage, err)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
	return msg, nil
}

// EditMessage invokes method messages.editMessage#48f71778 returning error if any. Edit message
func (ctx *Context) EditMessage(chatID int64, request *tg.MessagesEditMessageRequest) (*types.Message, error) {
	var err error
	if request == nil {
		request = &tg.MessagesEditMessageRequest{}
	}
	if request.Peer == nil {
		request.Peer, err = ctx.ResolveInputPeerByID(chatID)
		if err != nil {
			return nil, err
		}
	}

	upds, err := ctx.Raw.MessagesEditMessage(ctx, request)
	message, err := functions.ReturnEditMessageWithError(ctx.PeerStorage, upds, err)
	if err != nil {
		return nil, err
	}
	msg := types.ConstructMessageWithContext(message, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
	if ctx.setReply {
		_ = msg.SetRepliedToMessage(ctx.Context, ctx.Raw, ctx.PeerStorage)
	}
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

// GetChat returns tg.ChatFullClass of the provided chat id.
func (ctx *Context) GetChat(chatID int64) (tg.ChatFullClass, error) {
	inputPeer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return nil, err
	}
	switch p := inputPeer.(type) {
	case *tg.InputPeerChannel:
		channel, err := ctx.Raw.ChannelsGetFullChannel(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, err
		}
		return channel.FullChat, nil
	case *tg.InputPeerChat:
		chat, err := ctx.Raw.MessagesGetFullChat(ctx, chatID)
		if err != nil {
			return nil, err
		}
		return chat.FullChat, nil
	case *tg.InputPeerEmpty:
		return nil, mtp_errors.ErrPeerNotFound
	default:
		return nil, mtp_errors.ErrNotChat
	}
}

// GetUser returns tg.UserFull of the provided user id.
func (ctx *Context) GetUser(userID int64) (*tg.UserFull, error) {
	inputPeer, err := ctx.ResolveInputPeerByID(userID)
	if err != nil {
		return nil, err
	}
	switch p := inputPeer.(type) {
	case *tg.InputPeerUser:
		user, err := ctx.Raw.UsersGetFullUser(ctx, &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, err
		}
		return &user.FullUser, nil
	default:
		return nil, mtp_errors.ErrNotUser
	}
}

// GetChatMember fetches information about a chat member.
// For channels, returns tg.ChannelParticipantClass with member details.
// For regular chats, this is not yet implemented and returns an error.
//
// Parameters:
//   - chatID: The channel ID to query
//   - userID: The user ID to look up
//
// Returns the participant information or an error.
func (ctx *Context) GetChatMember(chatID, userID int64) (tg.ChannelParticipantClass, error) {
	peer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return nil, err
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		res, err := ctx.Raw.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
		})
		if err != nil {
			return nil, err
		}
		return res.Participant, nil

	case *tg.InputPeerChat, *tg.InputPeerUser:
		// For chats and users, not implemented yet
		return nil, fmt.Errorf("get chat member for chats not implemented yet")

	case *tg.InputPeerEmpty:
		return nil, mtp_errors.ErrPeerNotFound
	default:
		return nil, mtp_errors.ErrNotChat
	}
}

// GetMessages fetches messages from a chat by their IDs.
//
// Parameters:
//   - chatID: The chat ID containing the messages
//   - messageIDs: List of message identifiers (InputMessageID, InputMessageReplyTo, etc.)
//
// Returns the fetched messages or an error.
func (ctx *Context) GetMessages(chatID int64, messageIDs []tg.InputMessageClass) ([]tg.MessageClass, error) {
	return functions.GetMessages(ctx.Context, ctx.Raw, ctx.PeerStorage, chatID, messageIDs)
}

// BanChatMember bans a user from a chat until a specified date.
//
// Parameters:
//   - chatID: The chat ID (channel, group, or supergroup)
//   - userID: The user ID to ban
//   - untilDate: Unix timestamp for ban expiration (0 for permanent)
//
// Returns updates confirming the action or an error.
func (ctx *Context) BanChatMember(chatID, userID int64, untilDate int) (tg.UpdatesClass, error) {
	inputPeerChat, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return nil, err
	}
	switch inputPeerChat.(type) {
	case *tg.InputPeerChannel:
	case *tg.InputPeerChat:
	case *tg.InputPeerEmpty:
		return nil, mtp_errors.ErrPeerNotFound
	default:
		return nil, mtp_errors.ErrNotChat
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
		return nil, mtp_errors.ErrPeerNotFound
	default:
		return nil, mtp_errors.ErrNotUser
	}
	return functions.BanChatMember(ctx, ctx.Raw, inputPeerChat, inputPeerUser, untilDate)
}

// UnbanChatMember unbans a previously banned user from a channel.
//
// Parameters:
//   - chatID: The channel ID
//   - userID: The user ID to unban
//
// Returns true if successful, or an error.
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
		return false, mtp_errors.ErrPeerNotFound
	default:
		return false, mtp_errors.ErrNotChannel
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
		return false, mtp_errors.ErrPeerNotFound
	default:
		return false, mtp_errors.ErrNotUser
	}
	return functions.UnbanChatMember(ctx, ctx.Raw, inputPeerChat, inputPeerUser)
}

// AddChatMembers adds multiple users to a chat.
//
// Parameters:
//   - chatID: The chat ID (channel or group)
//   - userIDs: List of user IDs to add
//   - forwardLimit: Number of messages to forward from chat history (0-100)
//
// Returns true if successful, or an error.
func (ctx *Context) AddChatMembers(chatID int64, userIDs []int64, forwardLimit int) (bool, error) {
	inputPeerChat, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return false, err
	}
	switch inputPeerChat.(type) {
	case *tg.InputPeerChannel:
	case *tg.InputPeerChat:
	case *tg.InputPeerEmpty:
		return false, mtp_errors.ErrPeerNotFound
	default:
		return false, mtp_errors.ErrNotChat
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
			return false, mtp_errors.ErrPeerNotFound
		default:
			return false, mtp_errors.ErrNotUser
		}
	}
	return functions.AddChatMembers(ctx, ctx.Raw, inputPeerChat, userPeers, forwardLimit)
}

// ArchiveChats invokes method folders.editPeerFolders#6847d0ab returning error if any.
// Edit peers in peer folder¹
//
// Links:
//  1. https://core.telegram.org/api/folders#peer-folders
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
			return false, mtp_errors.ErrPeerNotFound
		default:
			return false, mtp_errors.ErrNotChat
		}
		chatPeers[i] = inputPeer
	}
	return functions.ArchiveChats(ctx, ctx.Raw, chatPeers)
}

// UnarchiveChats invokes method folders.editPeerFolders#6847d0ab returning error if any.
// Edit peers in peer folder¹
//
// Links:
//  1. https://core.telegram.org/api/folders#peer-folders
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
			return false, mtp_errors.ErrPeerNotFound
		default:
			return false, mtp_errors.ErrNotChat
		}
		chatPeers[i] = inputPeer
	}
	return functions.UnarchiveChats(ctx, ctx.Raw, chatPeers)
}

// CreateChannel invokes method channels.createChannel#3d5fb10f returning error if any.
// Create a supergroup/channel¹.
//
// Links:
//  1. https://core.telegram.org/api/channel
func (ctx *Context) CreateChannel(title, about string, broadcast bool) (*tg.Channel, error) {
	return functions.CreateChannel(ctx, ctx.Raw, ctx.PeerStorage, title, about, broadcast)
}

// CreateChat invokes method messages.createChat#9cb126e returning error if any. Creates a new chat.
func (ctx *Context) CreateChat(title string, userIDs []int64) (*tg.Chat, error) {
	userPeers := make([]tg.InputUserClass, len(userIDs))
	for i, uID := range userIDs {
		userPeer := ctx.ResolvePeerByID(uID)
		if userPeer.ID == 0 {
			return nil, mtp_errors.ErrPeerNotFound
		}
		if userPeer.Type != int(storage.TypeUser) {
			return nil, mtp_errors.ErrNotUser
		}
		userPeers[i] = &tg.InputUser{
			UserID:     userPeer.ID,
			AccessHash: userPeer.AccessHash,
		}
	}
	return functions.CreateChat(ctx, ctx.Raw, ctx.PeerStorage, title, userPeers)
}

// DeleteMessages shall be used to delete messages in a chat with chatID and messageIDs.
// Returns error if failed to delete.
func (ctx *Context) DeleteMessages(chatID int64, messageIDs []int) error {
	inputPeer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return err
	}
	switch p := inputPeer.(type) {
	case *tg.InputPeerChannel:
		_, err := ctx.Raw.ChannelsDeleteMessages(ctx, &tg.ChannelsDeleteMessagesRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			ID: messageIDs,
		})
		return err
	case *tg.InputPeerChat, *tg.InputPeerUser:
		_, err := ctx.Raw.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
			Revoke: true,
			ID:     messageIDs,
		})
		return err
	case *tg.InputPeerEmpty:
		return mtp_errors.ErrPeerNotFound
	default:
		return mtp_errors.ErrNotChat
	}
}

// ForwardMessage shall be used to forward messages in a chat with chatID and messageIDs.
// Returns updatesclass or an error if failed to delete.
//
// Deprecated: use ForwardMessages instead.
func (ctx *Context) ForwardMessage(fromChatID, toChatID int64, request *tg.MessagesForwardMessagesRequest) (tg.UpdatesClass, error) {
	return ctx.ForwardMessages(fromChatID, toChatID, request)
}

// ForwardMessages shall be used to forward messages in a chat with chatID and messageIDs.
// Returns updatesclass or an error if failed to delete.
func (ctx *Context) ForwardMessages(fromChatID, toChatID int64, request *tg.MessagesForwardMessagesRequest) (tg.UpdatesClass, error) {
	fromPeer, _ := ctx.ResolveInputPeerByID(fromChatID)
	if fromPeer.Zero() {
		return nil, fmt.Errorf("fromChatID: %w", mtp_errors.ErrPeerNotFound)
	}
	toPeer, _ := ctx.ResolveInputPeerByID(toChatID)
	if toPeer.Zero() {
		return nil, fmt.Errorf("toChatID: %w", mtp_errors.ErrPeerNotFound)
	}
	if request == nil {
		request = &tg.MessagesForwardMessagesRequest{}
	}
	request.RandomID = make([]int64, len(request.ID))
	for i := 0; i < len(request.ID); i++ {
		request.RandomID[i] = ctx.generateRandomID()
	}
	return ctx.Raw.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
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

type EditAdminOpts struct {
	AdminRights tg.ChatAdminRights
	AdminTitle  string
}

// PromoteChatMember is used to promote a user in a chat.
func (ctx *Context) PromoteChatMember(chatID, userID int64, opts *EditAdminOpts) (bool, error) {
	peerChat := ctx.ResolvePeerByID(chatID)
	if peerChat.ID == 0 {
		return false, fmt.Errorf("chat: %w", mtp_errors.ErrPeerNotFound)
	}
	peerUser := ctx.ResolvePeerByID(userID)
	if peerUser.ID == 0 {
		return false, fmt.Errorf("user: %w", mtp_errors.ErrPeerNotFound)
	}
	if opts == nil {
		opts = &EditAdminOpts{}
	}
	return functions.PromoteChatMember(ctx, ctx.Raw, peerChat, peerUser, opts.AdminRights, opts.AdminTitle)
}

// DemoteChatMember is used to demote a user in a chat.
func (ctx *Context) DemoteChatMember(chatID, userID int64, opts *EditAdminOpts) (bool, error) {
	peerChat := ctx.ResolvePeerByID(chatID)
	if peerChat.ID == 0 {
		return false, fmt.Errorf("chat: %w", mtp_errors.ErrPeerNotFound)
	}
	peerUser := ctx.ResolvePeerByID(userID)
	if peerUser.ID == 0 {
		return false, fmt.Errorf("user: %w", mtp_errors.ErrPeerNotFound)
	}
	if opts == nil {
		opts = &EditAdminOpts{}
	}
	return functions.DemoteChatMember(ctx, ctx.Raw, peerChat, peerUser, opts.AdminRights, opts.AdminTitle)
}

// ResolveUsername invokes method contacts.resolveUsername#f93ccba3 returning error if any.
// Resolve a @username to get peer info
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

// extractContactResolvedPeer converts a ContactsResolvedPeer response to an EffectiveChat.
// Used internally by ResolveUsername.
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

// GetUserProfilePhotos invokes method photos.getUserPhotos#91cd32a8 returning error if any. Returns the list of user photos.
func (ctx *Context) GetUserProfilePhotos(userID int64, opts *tg.PhotosGetUserPhotosRequest) ([]tg.PhotoClass, error) {
	peerUser := ctx.ResolvePeerByID(userID)
	if peerUser.ID == 0 {
		return nil, mtp_errors.ErrPeerNotFound
	}
	if opts == nil {
		opts = &tg.PhotosGetUserPhotosRequest{}
	}
	opts.UserID = &tg.InputUser{
		UserID:     userID,
		AccessHash: peerUser.AccessHash,
	}
	p, err := ctx.Raw.PhotosGetUserPhotos(ctx, opts)
	if err != nil {
		return nil, err
	}
	return p.GetPhotos(), nil
}

// ExportSessionString returns session of authorized account in the form of string.
// Note: This session string can be used to log back in with the help of gotg.
// Check session.SessionType for more information about it.
func (ctx *Context) ExportSessionString() (string, error) {
	return functions.EncodeSessionToString(ctx.PeerStorage.GetSession())
}

// DownloadOutputClass is an interface which is used to download media.
// It can be one from DownloadOutputStream, DownloadOutputPath and DownloadOutputParallel.
type DownloadOutputClass interface {
	run(context.Context, *downloader.Builder) (tg.StorageFileTypeClass, error)
}

// DownloadOutputStream is used to download media to an io.Writer.
// It can be used to download media to a file, memory etc.
type DownloadOutputStream struct {
	io.Writer
}

func (d DownloadOutputStream) run(ctx context.Context, b *downloader.Builder) (tg.StorageFileTypeClass, error) {
	return b.Stream(ctx, d)
}

// DownloadOutputPath is used to download media to a file.
type DownloadOutputPath string

func (d DownloadOutputPath) run(ctx context.Context, b *downloader.Builder) (tg.StorageFileTypeClass, error) {
	return b.ToPath(ctx, string(d))
}

// DownloadOutputParallel is used to download a media parallely.
type DownloadOutputParallel struct {
	io.WriterAt
}

func (d DownloadOutputParallel) run(ctx context.Context, b *downloader.Builder) (tg.StorageFileTypeClass, error) {
	return b.Parallel(ctx, d)
}

// DownloadMediaOpts object contains optional parameters for Context.DownloadMedia.
// If not provided, default values will be used.
type DownloadMediaOpts struct {
	// Threads sets downloading goroutines limit.
	Threads int
	//  If verify is true, file hashes will be checked. Verify is true by default for CDN downloads.
	Verify *bool
	// PartSize sets chunk size. Must be divisible by 4KB.
	PartSize int
}

// DownloadMedia downloads media from the provided MessageMediaClass.
// DownloadOutputClass can be one from DownloadOutputStream, DownloadOutputPath and DownloadOutputParallel.
// DownloadMediaOpts can be used to set optional parameters.
// Returns tg.StorageFileTypeClass and error if any.
func (ctx *Context) DownloadMedia(media tg.MessageMediaClass, downloadOutput DownloadOutputClass, opts *DownloadMediaOpts) (tg.StorageFileTypeClass, error) {
	if opts == nil {
		opts = &DownloadMediaOpts{}
	}
	mediaDownloader := downloader.NewDownloader()
	if opts.PartSize > 0 {
		mediaDownloader.WithPartSize(opts.PartSize)
	}
	inputFileLocation, err := functions.GetInputFileLocation(media)
	if err != nil {
		return nil, err
	}
	d := mediaDownloader.Download(ctx.Raw, inputFileLocation)
	if opts.Threads > 0 {
		d.WithThreads(opts.Threads)
	}
	if opts.Verify != nil {
		d.WithVerify(*opts.Verify)
	}
	return downloadOutput.run(ctx, d)
}

// TransferStarGift is used to transfer a star gift to a chat.
// Returns tg.UpdatesClass and error if any.
func (ctx *Context) TransferStarGift(chatID int64, starGift tg.InputSavedStarGiftClass) (tg.UpdatesClass, error) {
	peerUser, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return nil, mtp_errors.ErrPeerNotFound
	}
	upd, err := ctx.Raw.PaymentsTransferStarGift(ctx, &tg.PaymentsTransferStarGiftRequest{
		ToID:     peerUser,
		Stargift: starGift,
	})
	if err != nil {
		return nil, err
	}
	return upd, err
}

func (ctx *Context) ExportInvoice(inputMedia tg.InputMediaClass) (*tg.PaymentsExportedInvoice, error) {
	return ctx.Raw.PaymentsExportInvoice(ctx, inputMedia)
}

func (ctx *Context) SetPreCheckoutResults(success bool, queryID int64, err string) (bool, error) {
	return ctx.Raw.MessagesSetBotPrecheckoutResults(ctx, &tg.MessagesSetBotPrecheckoutResultsRequest{
		Success: success,
		QueryID: queryID,
		Error:   err,
	})
}

// ResolveInputPeerByID tries to resolve given id to InputPeer.
// Returns tg.InputPeerClass or error if peer could not be resolved
func (ctx *Context) ResolveInputPeerByID(id int64) (tg.InputPeerClass, error) {
	peerStorage := ctx.PeerStorage
	peer := peerStorage.GetInputPeerByID(id)
	if _, isEmpty := peer.(*tg.InputPeerEmpty); !isEmpty {
		return peer, nil
	}

	ID := constant.TDLibPeerID(id)
	if ID.IsChannel() {
		plainID := ID.ToPlain()
		chatsClass, err := ctx.Raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{
				ChannelID:  plainID,
				AccessHash: 0,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch channel: %w", err)
		}
		chat, ok := chatsClass.MapChats().First()
		if ok {
			if ch, ok := chat.(*tg.Channel); ok {
				peerStorage.AddPeer(plainID, ch.AccessHash, storage.TypeChannel, ch.Username)
				return ch.AsInputPeer(), nil
			}
		}
	} else if ID.IsChat() {
		plainID := ID.ToPlain()
		peer := &tg.InputPeerChat{
			ChatID: plainID,
		}
		peerStorage.AddPeer(plainID, storage.DefaultAccessHash, storage.TypeChat, storage.DefaultUsername)
		return peer, nil
	} else {
		plainID := ID.ToPlain()
		users, err := ctx.Raw.UsersGetUsers(ctx, []tg.InputUserClass{
			&tg.InputUser{
				UserID:     plainID,
				AccessHash: 0,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user: %w", err)
		}
		user, ok := tg.UserClassArray(users).FirstAsNotEmpty()
		if ok {
			peerStorage.AddPeer(plainID, user.AccessHash, storage.TypeUser, user.Username)
			return user.AsInputPeer(), nil
		}
		//Try to get from storage again, but this time with bot-api compatible ids
		if ID.IsUser() {
			ID.Channel(id)
			peer := peerStorage.GetInputPeerByID(int64(ID))
			if _, isEmpty := peer.(*tg.InputPeerEmpty); !isEmpty {
				return peer, nil
			}
			ID.Chat(id)
			peer = peerStorage.GetInputPeerByID(int64(ID))
			if _, isEmpty := peer.(*tg.InputPeerEmpty); !isEmpty {
				return peer, nil
			}
		}
	}

	return nil, mtp_errors.ErrPeerNotFound
}

// ResolvePeerByID tries to resolve given id to peer.
// ResolvePeerByID resolves a peer ID to a storage.Peer with metadata.
// Unlike ResolveInputPeerByID, this returns the stored peer information
// including ID, access hash, type, and username.
//
// Parameters:
//   - id: The peer ID to resolve
//
// Returns the Peer information or nil if not found.
func (ctx *Context) ResolvePeerByID(id int64) *storage.Peer {
	_, _ = ctx.ResolveInputPeerByID(id)
	peer := ctx.PeerStorage.GetPeerByID(id)
	if peer.ID != 0 {
		return peer
	}
	ID := constant.TDLibPeerID(id)
	if ID.IsUser() {
		ID.Channel(id)
		peer = ctx.ResolvePeerByID(int64(ID))
		if peer.ID != 0 {
			return peer
		}
		ID.Chat(id)
		peer = ctx.ResolvePeerByID(int64(ID))
		if peer.ID != 0 {
			return peer
		}

	}
	return peer
}

// PinMessage pins a message in a chat.
//
// Parameters:
//   - chatID: The chat ID containing the message
//   - messageID: The message ID to pin
//
// Returns updates confirming the action or an error.
func (ctx *Context) PinMessage(chatID int64, messageID int) (tg.UpdatesClass, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat ID cannot be empty")
	}
	if messageID == 0 {
		return nil, fmt.Errorf("message ID cannot be empty")
	}

	request := &tg.MessagesUpdatePinnedMessageRequest{
		ID: messageID,
	}

	peer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	if _, isEmpty := peer.(*tg.InputPeerEmpty); isEmpty {
		return nil, fmt.Errorf("peer not found: %d", chatID)
	}

	request.Peer = peer
	return ctx.Raw.MessagesUpdatePinnedMessage(ctx, request)
}

// UnpinMessage unpins a specific message in a chat.
//
// Parameters:
//   - chatID: The chat ID containing the message
//   - messageID: The message ID to unpin
//
// Returns an error if the operation fails.
func (ctx *Context) UnpinMessage(chatID int64, messageID int) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be empty")
	}
	if messageID == 0 {
		return fmt.Errorf("message ID cannot be empty")
	}

	peer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return fmt.Errorf("failed to resolve peer: %w", err)
	}

	if _, isEmpty := peer.(*tg.InputPeerEmpty); isEmpty {
		return fmt.Errorf("peer not found: %d", chatID)
	}

	_, err = ctx.Raw.MessagesUpdatePinnedMessage(ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:  peer,
		ID:    messageID,
		Unpin: true,
	})
	return err
}

// UnpinAllMessages unpins all messages in a chat.
//
// Parameters:
//   - chatID: The chat ID containing pinned messages
//
// Returns an error if the operation fails.
func (ctx *Context) UnpinAllMessages(chatID int64) error {
	if chatID == 0 {
		return fmt.Errorf("chat ID cannot be empty")
	}

	peer, err := ctx.ResolveInputPeerByID(chatID)
	if err != nil {
		return fmt.Errorf("failed to resolve peer: %w", err)
	}

	if _, isEmpty := peer.(*tg.InputPeerEmpty); isEmpty {
		return fmt.Errorf("peer not found: %d", chatID)
	}

	// Use MessagesUnpinAllMessages for all peer types
	_, err = ctx.Raw.MessagesUnpinAllMessages(ctx, &tg.MessagesUnpinAllMessagesRequest{
		Peer: peer,
	})
	return err
}
