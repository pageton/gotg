package adapter

import (
	"github.com/gotd/td/tg"
)

// SendOpts contains optional parameters for sending messages.
// Embeds tg.MessagesSendMessageRequest for direct access to all Telegram API fields.
type SendOpts struct {
	NoWebpage              bool
	Silent                 bool
	Background             bool
	ClearDraft             bool
	Noforwards             bool
	UpdateStickersetsOrder bool
	InvertMedia            bool
	AllowPaidFloodskip     bool
	Peer                   tg.InputPeerClass
	ReplyTo                tg.InputReplyToClass
	Message                string
	RandomID               int64
	ReplyMarkup            tg.ReplyMarkupClass
	Entities               []tg.MessageEntityClass
	ScheduleDate           int
	ScheduleRepeatPeriod   int
	SendAs                 tg.InputPeerClass
	QuickReplyShortcut     tg.InputQuickReplyShortcutClass
	Effect                 int64
	AllowPaidStars         int64
	SuggestedPost          tg.SuggestedPost
	ParseMode              string
	WithoutReply           bool
	ReplyMessageID         int
}

// SendMediaOpts contains optional parameters for sending media.
// Clones all fields from tg.MessagesSendMediaRequest.
type SendMediaOpts struct {
	Peer                   tg.InputPeerClass
	ReplyTo                tg.InputReplyToClass
	Media                  tg.InputMediaClass
	RandomID               int64
	ScheduleDate           int
	SendAs                 tg.InputPeerClass
	QuickReplyShortcut     tg.InputQuickReplyShortcutClass
	Effect                 int64
	AllowPaidFloodskip     bool
	Silent                 bool
	Background             bool
	ClearDraft             bool
	Noforwards             bool
	UpdateStickersetsOrder bool
	InvertMedia            bool
	AllowPaidStars         int64
	SuggestedPost          tg.SuggestedPost
	Message                string
	Entities               []tg.MessageEntityClass
	ReplyMarkup            tg.ReplyMarkupClass
	Caption                string
	ParseMode              string
	WithoutReply           bool
	ReplyMessageID         int
}

// EditOpts contains optional parameters for editing messages.
// Clones all fields from tg.MessagesEditMessageRequest.
type EditOpts struct {
	Flags                int
	NoWebpage            bool
	InvertMedia          bool
	Peer                 tg.InputPeerClass
	ID                   int
	Message              string
	Media                tg.InputMediaClass
	ReplyMarkup          tg.ReplyMarkupClass
	Entities             []tg.MessageEntityClass
	ScheduleDate         int
	ScheduleRepeatPeriod int
	QuickReplyShortcutID int
	ParseMode            string
}

// EditMediaOpts contains optional parameters for editing media messages.
// Clones all fields from tg.MessagesEditMessageRequest.
type EditMediaOpts struct {
	Flags                int
	NoWebpage            bool
	InvertMedia          bool
	Peer                 tg.InputPeerClass
	ID                   int
	Message              string
	Media                tg.InputMediaClass
	ReplyMarkup          tg.ReplyMarkupClass
	Entities             []tg.MessageEntityClass
	ScheduleDate         int
	ScheduleRepeatPeriod int
	QuickReplyShortcutID int
	Caption              string
	ParseMode            string
}
