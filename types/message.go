package types

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
)

// Message represents a Telegram message with bound context for method chaining.
// It wraps tg.Message and provides convenience methods for common operations
// like editing, replying, deleting, and accessing media.
type Message struct {
	*tg.Message
	ReplyToMessage *Message
	Text           string
	IsService      bool
	Action         tg.MessageActionClass
	// Context fields for bound methods
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// ConstructMessage creates a Message from a tg.MessageClass without context binding.
// The returned Message will have nil context fields and cannot perform bound operations.
// For full functionality, use ConstructMessageWithContext instead.
func ConstructMessage(m tg.MessageClass) *Message {
	return ConstructMessageWithContext(m, nil, nil, nil, 0)
}

// ConstructMessageWithContext creates a Message from a tg.MessageClass with full context binding.
// This is the preferred constructor as it enables bound methods like Edit(), Reply(), etc.
//
// Parameters:
//   - m: The message class (Message, MessageService, or MessageEmpty)
//   - ctx: Context for cancellation and timeouts
//   - raw: The raw Telegram client for API calls
//   - peerStorage: Peer storage for resolving peer references
//   - selfID: The current user/bot ID for context
func ConstructMessageWithContext(m tg.MessageClass, ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, selfID int64) *Message {
	switch msg := m.(type) {
	case *tg.Message:
		return constructMessageFromMessageWithContext(msg, ctx, raw, peerStorage, selfID)
	case *tg.MessageService:
		return constructMessageFromMessageServiceWithContext(msg, ctx, raw, peerStorage, selfID)
	case *tg.MessageEmpty:
		return constructMessageFromMessageEmptyWithContext(msg, ctx, raw, peerStorage, selfID)
	}
	return &Message{Ctx: ctx, RawClient: raw, PeerStorage: peerStorage, SelfID: selfID}
}

// constructMessageFromMessageWithContext creates a Message from tg.Message with context binding.
func constructMessageFromMessageWithContext(m *tg.Message, ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, selfID int64) *Message {
	return &Message{
		Message:     m,
		Text:        m.Message,
		Ctx:         ctx,
		RawClient:   raw,
		PeerStorage: peerStorage,
		SelfID:      selfID,
	}
}

// constructMessageFromMessageEmptyWithContext creates a Message from tg.MessageEmpty with context binding.
func constructMessageFromMessageEmptyWithContext(m *tg.MessageEmpty, ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, selfID int64) *Message {
	return &Message{
		Message: &tg.Message{
			ID:     m.ID,
			PeerID: m.PeerID,
		},
		Ctx:         ctx,
		RawClient:   raw,
		PeerStorage: peerStorage,
		SelfID:      selfID,
	}
}

// constructMessageFromMessageServiceWithContext creates a Message from tg.MessageService with context binding.
// Service messages represent actions like chat creation, member joins, etc.
func constructMessageFromMessageServiceWithContext(m *tg.MessageService, ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, selfID int64) *Message {
	return &Message{
		Message: &tg.Message{
			Out:         m.Out,
			Mentioned:   m.Mentioned,
			MediaUnread: m.MediaUnread,
			Silent:      m.Silent,
			Post:        m.Post,
			Legacy:      m.Legacy,
			ID:          m.ID,
			Date:        m.Date,
			FromID:      m.FromID,
			PeerID:      m.PeerID,
			ReplyTo:     m.ReplyTo,
			TTLPeriod:   m.TTLPeriod,
		},
		IsService:   true,
		Action:      m.Action,
		Ctx:         ctx,
		RawClient:   raw,
		PeerStorage: peerStorage,
		SelfID:      selfID,
	}
}

// SetRepliedToMessage populates the ReplyToMessage field by fetching the message being replied to.
// This is a lazy-loading operation that fetches the reply message from Telegram.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//
// Returns an error if:
//   - The message is not a reply (ReplyTo is not MessageReplyHeader)
//   - The replied message no longer exists
//   - The API call fails
func (m *Message) SetRepliedToMessage(ctx context.Context, raw *tg.Client, p *storage.PeerStorage) error {
	replyMessage, ok := m.ReplyTo.(*tg.MessageReplyHeader)
	if !ok {
		return errors.ErrReplyNotMessage
	}
	replyTo := replyMessage.ReplyToMsgID
	if replyTo == 0 {
		return errors.ErrMessageNotExist
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	msgs, err := functions.GetMessages(ctx, raw, p, chatID, []tg.InputMessageClass{
		&tg.InputMessageID{
			ID: replyTo,
		},
	})
	if err != nil {
		return err
	}
	m.ReplyToMessage = ConstructMessageWithContext(msgs[0], ctx, raw, p, m.SelfID)
	return nil
}

// Delete deletes the message.
// Returns error if failed to delete.
func (m *Message) Delete() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	return functions.DeleteMessages(m.Ctx, m.RawClient, m.PeerStorage, chatID, []int{m.ID})
}

// IsOutgoing returns true if the message was sent by this client (self).
// It checks the native tg.Message.Out flag first, then falls back to
// comparing the sender's user ID with SelfID. This fallback is necessary
// because bot accounts always receive Out=false from Telegram's MTProto.
func (m *Message) IsOutgoing() bool {
	if m == nil || m.Message == nil {
		return false
	}
	if m.Out {
		return true
	}
	if m.SelfID == 0 {
		return false
	}
	// For bot accounts, Out is always false. Fall back to sender comparison.
	if uid := m.UserID(); uid != 0 {
		return uid == m.SelfID
	}
	// FromID nil in private chats means the message is from the session owner.
	if m.FromID == nil {
		if peerUser, ok := m.PeerID.(*tg.PeerUser); ok {
			return peerUser.UserID == m.SelfID
		}
	}
	return false
}

// UserID returns the user ID of the message sender.
// Returns 0 if the message has no sender or sender is not a user.
func (m *Message) UserID() int64 {
	if m == nil || m.Message == nil {
		return 0
	}
	if m.FromID == nil {
		return 0
	}
	peerUser, ok := m.FromID.(*tg.PeerUser)
	if !ok {
		return 0
	}
	return peerUser.UserID
}
