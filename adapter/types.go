package adapter

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/conv"
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
	// original context of client.
	context.Context

	setReply    bool
	random      *rand.Rand
	PeerStorage *storage.PeerStorage
	Conv        *conv.Manager
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
	// InlineQuery is the tg.UpdateInlineBotCallbackQuery of current update.
	InlineQuery *tg.UpdateBotInlineQuery
	// ChatJoinRequest is the tg.UpdatePendingJoinRequests of current update.
	ChatJoinRequest *tg.UpdatePendingJoinRequests
	// ChatParticipant is the tg.UpdateChatParticipant of current update.
	ChatParticipant *tg.UpdateChatParticipant
	// ChannelParticipant is the tg.UpdateChannelParticipant of current update.
	ChannelParticipant *tg.UpdateChannelParticipant
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
}

// NewContext creates a new Context object with provided parameters.
func NewContext(ctx context.Context, client *tg.Client, peerStorage *storage.PeerStorage, self *tg.User, sender *message.Sender, entities *tg.Entities, setReply bool, conv *conv.Manager) *Context {
	return &Context{
		Context:     ctx,
		Raw:         client,
		Self:        self,
		Sender:      sender,
		Entities:    entities,
		random:      rand.New(rand.NewSource(time.Now().UnixNano())),
		setReply:    setReply,
		PeerStorage: peerStorage,
		Conv:        conv,
	}
}

// GetNewUpdate creates a new Update with provided parameters.
func GetNewUpdate(ctx *Context, update tg.UpdateClass) *Update {
	u := &Update{
		UpdateClass: update,
		Entities:    ctx.Entities,
		Ctx:         ctx,
		Self:        ctx.Self,
	}
	switch update := update.(type) {
	case *tg.UpdateNewMessage:
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		diff, err := ctx.Raw.UpdatesGetDifference(ctx.Context, &tg.UpdatesGetDifferenceRequest{
			Pts:  update.Pts - 1,
			Date: int(time.Now().Unix()),
		})
		// Silently add caught entities to *tg.Entities
		if err == nil {
			if value, ok := diff.(*tg.UpdatesDifference); ok {
				for _, vu := range value.Chats {
					switch chat := vu.(type) {
					case *tg.Chat:
						ctx.Entities.Chats[chat.ID] = chat
						if ctx.PeerStorage.GetPeerByID(chat.ID) != nil {
							continue
						}
						ctx.PeerStorage.AddPeer(chat.ID, storage.DefaultAccessHash, storage.TypeChat, storage.DefaultUsername)
					case *tg.Channel:
						ctx.Entities.Channels[chat.ID] = chat
						if chat.Min || ctx.PeerStorage.GetPeerByID(chat.ID) != nil {
							continue
						}
						ctx.PeerStorage.AddPeer(chat.ID, chat.AccessHash, storage.TypeChannel, chat.Username)
					}
				}
				for _, vu := range value.Users {
					user, ok := vu.AsNotEmpty()
					if !ok {
						continue
					}
					ctx.Entities.Users[user.ID] = user
					if user.Min || ctx.PeerStorage.GetPeerByID(user.ID) != nil {
						continue
					}
					ctx.PeerStorage.AddPeer(user.ID, user.AccessHash, storage.TypeUser, user.Username)
				}
			}
		}
		u.fillUserIDFromMessage(ctx.Self.ID)
	case message.AnswerableMessageUpdate:
		m := update.GetMessage()
		u.EffectiveMessage = types.ConstructMessageWithContext(m, ctx.Context, ctx.Raw, ctx.PeerStorage, ctx.Self.ID)
		u.fillUserIDFromMessage(ctx.Self.ID)
	case *tg.UpdateBotCallbackQuery:
		u.CallbackQuery = update
		u.userID = update.UserID
	case *tg.UpdateBotInlineQuery:
		u.InlineQuery = update
		u.userID = update.UserID
	case *tg.UpdatePendingJoinRequests:
		u.ChatJoinRequest = update
	case *tg.UpdateChatParticipant:
		u.ChatParticipant = update
		u.userID = update.UserID
	case *tg.UpdateChannelParticipant:
		u.ChannelParticipant = update
		u.userID = update.UserID
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

// FormatHelper provides text formatting convenience methods.
// Use Update.Format() to get a formatter instance.
type FormatHelper struct {
	mode string // HTML, Markdown, or ""
}

// Format returns a new FormatHelper for formatting text.
// The mode determines the formatting style: "HTML", "Markdown", "MarkdownV2", or "" (none).
func (u *Update) Format(mode string) *FormatHelper {
	return &FormatHelper{mode: mode}
}

// Bold wraps text with bold formatting markers.
func (f *FormatHelper) Bold(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<b>%s</b>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("*%s*", escapeMarkdownV2(text))
	}
	return text
}

// Italic wraps text with italic formatting markers.
func (f *FormatHelper) Italic(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<i>%s</i>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("_%s_", escapeMarkdownV2(text))
	}
	return text
}

// Underline wraps text with underline formatting markers.
func (f *FormatHelper) Underline(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<u>%s</u>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("__%s__", escapeMarkdownV2(text))
	}
	return text
}

// Strikethrough wraps text with strikethrough formatting markers.
func (f *FormatHelper) Strikethrough(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<s>%s</s>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("~%s~", escapeMarkdownV2(text))
	}
	return text
}

// Spoiler wraps text with spoiler formatting markers.
func (f *FormatHelper) Spoiler(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<tg-spoiler>%s</tg-spoiler>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("||%s||", escapeMarkdownV2(text))
	}
	return text
}

// Code wraps text with code formatting markers.
func (f *FormatHelper) Code(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<code>%s</code>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("`%s`", text)
	}
	return text
}

// Pre wraps text with pre-formatted code block markers.
func (f *FormatHelper) Pre(text string) string {
	return f.PreWithLanguage(text, "")
}

// PreWithLanguage wraps text with pre-formatted code block markers with syntax highlighting.
func (f *FormatHelper) PreWithLanguage(text, language string) string {
	switch f.mode {
	case HTML:
		if language != "" {
			return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", language, text)
		}
		return fmt.Sprintf("<pre>%s</pre>", text)
	case Markdown, MarkdownV2:
		if language != "" {
			return fmt.Sprintf("```%s\n%s\n```", language, text)
		}
		return fmt.Sprintf("```\n%s\n```", text)
	}
	return text
}

// Link creates a hyperlink with the specified URL.
func (f *FormatHelper) Link(text, url string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<a href='%s'>%s</a>", url, text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf("[%s](%s)", escapeMarkdownV2(text), url)
	}
	return fmt.Sprintf("%s: %s", text, url)
}

// Mention creates a mention link for a Telegram user.
func (f *FormatHelper) Mention(displayName string, userID int64) string {
	link := fmt.Sprintf("tg://user?id=%d", userID)
	return f.Link(displayName, link)
}

// CustomEmoji creates a custom emoji link.
func (f *FormatHelper) CustomEmoji(emoji string, emojiID int64) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<tg-emoji emoji-id=\"%d\">%s</tg-emoji>", emojiID, emoji)
	case Markdown, MarkdownV2:
		link := fmt.Sprintf("tg://emoji?id=%d", emojiID)
		return fmt.Sprintf("[%s](%s)", emoji, link)
	}
	return emoji
}

// Blockquote wraps text with blockquote formatting markers.
func (f *FormatHelper) Blockquote(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<blockquote>%s</blockquote>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf(">%s", text)
	}
	return text
}

// ExpandableBlockquote wraps text with expandable blockquote formatting markers.
func (f *FormatHelper) ExpandableBlockquote(text string) string {
	switch f.mode {
	case HTML:
		return fmt.Sprintf("<blockquote expandable>%s</blockquote>", text)
	case Markdown, MarkdownV2:
		return fmt.Sprintf(">%s||", text)
	}
	return text
}

// escapeMarkdownV2 escapes special characters for MarkdownV2.
func escapeMarkdownV2(text string) string {
	escapeMap := map[rune]string{
		'_': "\\_",
		'*': "\\*",
		'[': "\\[",
		']': "\\]",
		'(': "\\(",
		')': "\\)",
		'~': "\\~",
		'`': "\\`",
		'>': "\\>",
		'#': "\\#",
		'+': "\\+",
		'-': "\\-",
		'=': "\\=",
		'|': "\\|",
		'{': "\\{",
		'}': "\\}",
		'.': "\\.",
		'!': "\\!",
	}
	var result strings.Builder
	for _, r := range text {
		if escaped, ok := escapeMap[r]; ok {
			result.WriteString(escaped)
		} else {
			result.WriteString(string(r))
		}
	}
	return result.String()
}
