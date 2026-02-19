package adapter

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/conv"
	"github.com/pageton/gotg/log"
	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/types"
)

// Parse mode constants
const (
	HTML       = "HTML"
	Markdown   = "Markdown"
	MarkdownV2 = "MarkdownV2"
	ModeNone   = ""
)

// Context consists of context.Context, tg.Client, Self etc.
type Context struct {
	// raw tg client which will be used make send requests.
	Raw *tg.Client
	// self user who authorized the session.
	Self *tg.User
	// Sender is a message sending helper.
	Sender *message.Sender
	// Entities consists of mapped users, chats and channels from the update.
	Entities *tg.Entities
	// TelegramClient is the high-level telegram.Client used for DC-routed operations
	// such as editing inline messages on the correct DC.
	TelegramClient *telegram.Client
	// GetDCPool returns a cached tg.Invoker for the given DC ID.
	// Pools are created once and reused for the bot's lifetime.
	GetDCPool func(ctx context.Context, dcID int) (tg.Invoker, error)
	// original context of client.
	context.Context

	setReply         bool
	DefaultParseMode string
	PeerStorage      *storage.PeerStorage
	Conv             *conv.Manager
	OnOutgoing       func(*FakeOutgoingUpdate)
	// Logger is the logger instance passed from the dispatcher.
	Logger *log.Logger
	// Translator for i18n support
	Translator interface {
		Get(userID int64, key string, args ...any) string
		GetCtx(userID int64, key string, ctx any) string
		SetLang(userID int64, lang any)
		GetLang(userID int64) any
	}
}

// Update contains all data related to an update.
type Update struct {
	// EffectiveMessage is the tg.Message of current update.
	EffectiveMessage *types.Message
	// CallbackQuery is the tg.UpdateBotCallbackQuery of current update.
	CallbackQuery *tg.UpdateBotCallbackQuery
	// InlineCallbackQuery is the tg.UpdateInlineBotCallbackQuery of current update (for inline message buttons).
	InlineCallbackQuery *tg.UpdateInlineBotCallbackQuery
	// InlineQuery is the wrapped inline query of current update.
	InlineQuery *types.InlineQuery
	// ChatJoinRequest is the tg.UpdatePendingJoinRequests of current update.
	ChatJoinRequest *tg.UpdatePendingJoinRequests
	// ChatParticipant is the tg.UpdateChatParticipant of current update.
	ChatParticipant *tg.UpdateChatParticipant
	// ChannelParticipant is the tg.UpdateChannelParticipant of current update.
	ChannelParticipant *tg.UpdateChannelParticipant
	// ChosenInlineResult is the wrapped chosen inline result of current update.
	ChosenInlineResult *types.ChosenInlineResult
	// DeletedMessages is the tg.UpdateDeleteMessages of current update.
	DeletedMessages *tg.UpdateDeleteMessages
	// DeletedChannelMessages is the tg.UpdateDeleteChannelMessages of current update.
	DeletedChannelMessages *tg.UpdateDeleteChannelMessages
	// MessageReaction is the tg.UpdateBotMessageReaction of current update.
	MessageReaction *tg.UpdateBotMessageReaction
	// ChatBoost is the tg.UpdateBotChatBoost of current update.
	ChatBoost *tg.UpdateBotChatBoost
	// BusinessConnection is the tg.UpdateBotBusinessConnect of current update.
	BusinessConnection *tg.UpdateBotBusinessConnect
	// BusinessMessage is the tg.UpdateBotNewBusinessMessage of current update.
	BusinessMessage *tg.UpdateBotNewBusinessMessage
	// BusinessEditedMessage is the tg.UpdateBotEditBusinessMessage of current update.
	BusinessEditedMessage *tg.UpdateBotEditBusinessMessage
	// BusinessDeletedMessages is the tg.UpdateBotDeleteBusinessMessage of current update.
	BusinessDeletedMessages *tg.UpdateBotDeleteBusinessMessage
	// BusinessCallbackQuery is the tg.UpdateBusinessBotCallbackQuery of current update.
	BusinessCallbackQuery *tg.UpdateBusinessBotCallbackQuery
	// EffectiveOutgoing contains metadata for synthetic outgoing updates (send/edit/delete).
	EffectiveOutgoing *FakeOutgoingUpdate
	// Log is the logger for this update. Use u.Log.Info("msg", "key", val) etc.
	Log *log.Logger
	// UpdateClass is the current update in raw form.
	UpdateClass tg.UpdateClass
	// Entities of an update, i.e. mapped users, chats and channels.
	Entities *tg.Entities
	// User id of the user responsible for the update.
	userID int64
	// Sender is a message sending helper.
	Self *tg.User
	// Context of the current update.
	Ctx *Context
	// IsEdited indicates the effective message was an edit, not a new message.
	IsEdited bool
}

// NewContext creates a new Context object with provided parameters.
func NewContext(ctx context.Context, client *tg.Client, peerStorage *storage.PeerStorage, self *tg.User, sender *message.Sender, entities *tg.Entities, setReply bool, conv *conv.Manager, logger *log.Logger, telegramClient ...*telegram.Client) *Context {
	c := &Context{
		Context:     ctx,
		Raw:         client,
		Self:        self,
		Sender:      sender,
		Entities:    entities,
		setReply:    setReply,
		PeerStorage: peerStorage,
		Conv:        conv,
		Logger:      logger,
	}
	if len(telegramClient) > 0 {
		c.TelegramClient = telegramClient[0]
	}
	return c
}

// GetNewUpdate creates a new Update with provided parameters.
func GetNewUpdate(ctx *Context, update tg.UpdateClass) *Update {
	u := &Update{
		UpdateClass: update,
		Entities:    ctx.Entities,
		Ctx:         ctx,
		Self:        ctx.Self,
		Log:         ctx.Logger,
	}
	switch update := update.(type) {
	case *tg.UpdateNewMessage:
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.fillUserIDFromMessage(ctx.Self.ID)
	case *tg.UpdateEditMessage:
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.IsEdited = true
		u.fillUserIDFromMessage(ctx.Self.ID)
	case *tg.UpdateEditChannelMessage:
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.IsEdited = true
		u.fillUserIDFromMessage(ctx.Self.ID)
	case *tg.UpdateBotCallbackQuery:
		u.CallbackQuery = update
		u.userID = update.UserID
	case *tg.UpdateInlineBotCallbackQuery:
		u.InlineCallbackQuery = update
		u.userID = update.UserID
	case *tg.UpdateBotInlineQuery:
		u.InlineQuery = types.ConstructInlineQueryWithContext(update, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID, ctx.Entities)
		u.userID = update.UserID
	case *tg.UpdateBotInlineSend:
		u.ChosenInlineResult = types.ConstructChosenInlineResultWithContext(update, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID, ctx.Entities)
		u.userID = update.UserID
	case *tg.UpdatePendingJoinRequests:
		u.ChatJoinRequest = update
	case *tg.UpdateChatParticipant:
		u.ChatParticipant = update
		u.userID = update.UserID
	case *tg.UpdateChannelParticipant:
		u.ChannelParticipant = update
		u.userID = update.UserID
	case *tg.UpdateDeleteMessages:
		u.DeletedMessages = update
	case *tg.UpdateDeleteChannelMessages:
		u.DeletedChannelMessages = update
	case *tg.UpdateBotMessageReaction:
		u.MessageReaction = update
	case *tg.UpdateBotChatBoost:
		u.ChatBoost = update
	case *tg.UpdateBotBusinessConnect:
		u.BusinessConnection = update
	case *tg.UpdateBotNewBusinessMessage:
		u.BusinessMessage = update
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.fillUserIDFromMessage(ctx.Self.ID)
	case *tg.UpdateBotEditBusinessMessage:
		u.BusinessEditedMessage = update
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.IsEdited = true
		u.fillUserIDFromMessage(ctx.Self.ID)
	case *tg.UpdateBotDeleteBusinessMessage:
		u.BusinessDeletedMessages = update
	case *tg.UpdateBusinessBotCallbackQuery:
		u.BusinessCallbackQuery = update
		u.userID = update.UserID
	case message.AnswerableMessageUpdate:
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.fillUserIDFromMessage(ctx.Self.ID)
	}
	if u.EffectiveMessage == nil {
		u.EffectiveMessage = &types.Message{Message: &tg.Message{}}
	}
	return u
}

// CallbackOptions contains optional parameters for answering callback queries.
type CallbackOptions struct {
	// Alert shows the answer as a popup alert instead of a toast notification.
	Alert bool
	// CacheTime is the time in seconds for which the result of a callback query
	// may be cached on the client side.
	CacheTime int
	// URL is the URL to open as a game or in-browser app.
	URL string
}
