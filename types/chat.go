package types

import (
	"context"
	"fmt"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
)

// EffectiveChat interface covers the all three types of chats:
// - tg.User
// - tg.Chat
// - tg.Channel
//
// This interface is implemented by the following structs:
// - User: If the chat is a tg.User then this struct will be returned.
// - Chat: if the chat is a tg.Chat then this struct will be returned.
// - Channel: if the chat is a tg.Channel then this struct will be returned.
// - EmptyUC: if the PeerID doesn't match any of the above cases then EmptyUC struct is returned.
type EffectiveChat interface {
	GetID() int64
	GetAccessHash() int64
	IsChannel() bool
	IsChat() bool
	IsUser() bool
	GetInputUser() tg.InputUserClass
	GetInputChannel() tg.InputChannelClass
	GetInputPeer() tg.InputPeerClass
}

// Participant represents a chat/channel member with their details.
type Participant struct {
	User        *User
	Participant tg.ChannelParticipantClass
	Status      string
	Rights      *tg.ChatAdminRights
	Title       string
	UserID      int64
	ChatID      int64
}

// EmptyUC implements EffectiveChat interface for empty chats.
type EmptyUC struct{}

// GetID returns the chat ID.
// Always returns 0 for EmptyUC.
func (*EmptyUC) GetID() int64 {
	return 0
}

// GetAccessHash returns the access hash for API authentication.
// Always returns 0 for EmptyUC.
func (*EmptyUC) GetAccessHash() int64 {
	return 0
}

// GetInputUser returns the InputUser for this peer.
// Always returns nil for EmptyUC.
func (*EmptyUC) GetInputUser() tg.InputUserClass {
	return nil
}

// GetInputChannel returns the InputChannel for this peer.
// Always returns nil for EmptyUC.
func (*EmptyUC) GetInputChannel() tg.InputChannelClass {
	return nil
}

// GetInputPeer returns the InputPeer for this peer.
// Always returns nil for EmptyUC.
func (*EmptyUC) GetInputPeer() tg.InputPeerClass {
	return nil
}

// IsChannel returns true if this effective chat is a channel.
// Always returns false for EmptyUC.
func (*EmptyUC) IsChannel() bool {
	return false
}

// IsChat returns true if this effective chat is a group chat.
// Always returns false for EmptyUC.
func (*EmptyUC) IsChat() bool {
	return false
}

// IsUser returns true if this effective chat is a user (private chat).
// Always returns false for EmptyUC.
func (*EmptyUC) IsUser() bool {
	return false
}

// Chat implements EffectiveChat interface for tg.Chat chats.
type Chat struct {
	tg.Chat
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// GetID returns the chat ID in TDLib format.
// The ID is encoded with the chat prefix for proper peer identification.
func (u *Chat) GetID() int64 {
	var ID constant.TDLibPeerID
	ID.Chat(u.ID)
	return int64(ID)
}

// GetAccessHash returns the access hash for API authentication.
// Always returns 0 for Chat (group chats don't use access hashes).
func (*Chat) GetAccessHash() int64 {
	return 0
}

// GetInputUser returns the InputUser for this peer.
// Always returns nil for Chat (chats are not users).
func (*Chat) GetInputUser() tg.InputUserClass {
	return nil
}

// GetInputChannel returns the InputChannel for this peer.
// Always returns nil for Chat (chats are not channels).
func (*Chat) GetInputChannel() tg.InputChannelClass {
	return nil
}

// GetInputPeer returns the InputPeer for API calls.
func (v *Chat) GetInputPeer() tg.InputPeerClass {
	return &tg.InputPeerChat{
		ChatID: v.ID,
	}
}

// IsChannel returns true if this effective chat is a channel.
// Always returns false for Chat.
func (*Chat) IsChannel() bool {
	return false
}

// IsChat returns true if this effective chat is a group chat.
// Always returns true for Chat.
func (*Chat) IsChat() bool {
	return true
}

// IsUser returns true if this effective chat is a user (private chat).
// Always returns false for Chat.
func (*Chat) IsUser() bool {
	return false
}

// Raw returns the underlying tg.Chat struct.
func (c *Chat) Raw() *tg.Chat {
	return &c.Chat
}

// GetChat returns the basic chat information for this chat.
// Returns tg.ChatClass which can be type-asserted to *tg.Chat.
func (c *Chat) GetChat() (tg.ChatClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("chat has no client context")
	}
	return functions.GetChat(c.Ctx, c.RawClient, c.PeerStorage, c.GetID())
}

// GetChatInviteLink generates an invite link for this chat.
//
// Parameters:
//   - req: Telegram's MessagesExportChatInviteRequest (use &tg.MessagesExportChatInviteRequest{} for default)
//
// Returns exported chat invite or an error.
func (c *Chat) GetChatInviteLink(req ...*tg.MessagesExportChatInviteRequest) (*tg.ExportedChatInviteClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("chat has no client context")
	}
	peer := c.PeerStorage.GetInputPeerByID(c.GetID())
	if peer == nil {
		peer = &tg.InputPeerChat{
			ChatID: c.ID,
		}
	}

	if len(req) == 0 || req[0] == nil {
		req = []*tg.MessagesExportChatInviteRequest{{}}
	}

	link, err := c.RawClient.MessagesExportChatInvite(c.Ctx, &tg.MessagesExportChatInviteRequest{
		Peer:                  peer,
		LegacyRevokePermanent: req[0].LegacyRevokePermanent,
		RequestNeeded:         req[0].RequestNeeded,
		UsageLimit:            req[0].UsageLimit,
		Title:                 req[0].Title,
		ExpireDate:            req[0].ExpireDate,
	})
	if err != nil {
		return nil, err
	}
	return &link, nil
}
