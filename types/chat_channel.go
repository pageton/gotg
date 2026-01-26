package types

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/storage"
)

// Channel implements EffectiveChat interface for tg.Channel chats.
type Channel struct {
	tg.Channel
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// GetID returns the channel ID in TDLib format.
// The ID is encoded with the channel prefix for proper peer identification.
func (u *Channel) GetID() int64 {
	var ID constant.TDLibPeerID
	ID.Channel(u.ID)
	return int64(ID)
}

// GetAccessHash returns the access hash for API authentication.
func (u *Channel) GetAccessHash() int64 {
	return u.AccessHash
}

// GetInputUser returns the InputUser for this peer.
// Always returns nil for Channel (channels are not users).
func (*Channel) GetInputUser() tg.InputUserClass {
	return nil
}

// GetInputChannel returns the InputChannel for API calls.
func (v *Channel) GetInputChannel() tg.InputChannelClass {
	return &tg.InputChannel{
		ChannelID:  v.ID,
		AccessHash: v.AccessHash,
	}
}

// GetInputPeer returns the InputPeer for API calls.
func (v *Channel) GetInputPeer() tg.InputPeerClass {
	return &tg.InputPeerChannel{
		ChannelID:  v.ID,
		AccessHash: v.AccessHash,
	}
}

// IsChannel returns true if this effective chat is a channel.
// Always returns true for Channel.
func (*Channel) IsChannel() bool {
	return true
}

// IsChat returns true if this effective chat is a group chat.
// Always returns false for Channel.
func (*Channel) IsChat() bool {
	return false
}

// IsUser returns true if this effective chat is a user (private chat).
// Always returns false for Channel.
func (*Channel) IsUser() bool {
	return false
}

// Raw returns the underlying tg.Channel struct.
func (c *Channel) Raw() *tg.Channel {
	return &c.Channel
}

// GetChat returns the basic chat information for this channel.
// Returns tg.ChatClass which can be type-asserted to *tg.Channel.
func (c *Channel) GetChat() (tg.ChatClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("channel has no client context")
	}
	return functions.GetChat(c.Ctx, c.RawClient, c.PeerStorage, c.GetID())
}

// GetChatInviteLink generates an invite link for this channel.
//
// Parameters:
//   - req: Telegram's MessagesExportChatInviteRequest (use &tg.MessagesExportChatInviteRequest{} for default)
//
// Returns exported chat invite or an error.
func (c *Channel) GetChatInviteLink(req ...*tg.MessagesExportChatInviteRequest) (*tg.ExportedChatInviteClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("channel has no client context")
	}
	peer := c.PeerStorage.GetInputPeerByID(c.GetID())
	if peer == nil {
		peer = &tg.InputPeerChannel{
			ChannelID:  c.ID,
			AccessHash: c.AccessHash,
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

// GetChatMember fetches information about a specific member of this channel.
// Returns tg.ChannelParticipantClass with the member information.
func (c *Channel) GetChatMember(userID int64) (tg.ChannelParticipantClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("channel has no client context")
	}

	peer := c.PeerStorage.GetInputPeerByID(userID)
	if peer == nil {
		peer = &tg.InputPeerUser{UserID: userID}
	}

	res, err := c.RawClient.ChannelsGetParticipant(c.Ctx, &tg.ChannelsGetParticipantRequest{
		Channel: &tg.InputChannel{
			ChannelID:  c.ID,
			AccessHash: c.AccessHash,
		},
		Participant: peer,
	})
	if err != nil {
		return nil, err
	}
	return res.Participant, nil
}

// SendMessage sends a message to this channel.
// Opts can be:
//   - string (parse mode): "HTML", "Markdown", "MarkdownV2", or ""
//   - []tg.MessageEntityClass (raw entities for backward compatibility)
//   - struct with fields: ParseMode, NoWebpage, Silent, Markup
func (c *Channel) SendMessage(text string, opts ...any) (*Message, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("channel has no client context")
	}
	peer := c.PeerStorage.GetInputPeerByID(c.GetID())
	if peer == nil {
		peer = &tg.InputPeerChannel{
			ChannelID:  c.ID,
			AccessHash: c.AccessHash,
		}
	}

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var noWebpage, silent bool
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		if s, ok := opts[0].(string); ok {
			parseMode = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			ents = e
		} else {
			val := reflect.ValueOf(opts[0])
			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				for i := 0; i < val.NumField(); i++ {
					field := val.Type().Field(i)
					value := val.Field(i)
					if !value.IsValid() {
						continue
					}
					switch field.Name {
					case "Entities":
						if e, ok := value.Interface().([]tg.MessageEntityClass); ok {
							ents = e
						}
					case "Markup":
						if rm, ok := value.Interface().(tg.ReplyMarkupClass); ok {
							markup = rm
						}
					case "NoWebpage":
						if nw, ok := value.Interface().(bool); ok {
							noWebpage = nw
						}
					case "Silent":
						if s, ok := value.Interface().(bool); ok {
							silent = s
						}
					case "ParseMode":
						if s, ok := value.Interface().(string); ok {
							parseMode = s
						}
					}
				}
			}
		}
	}

	if len(ents) == 0 && parseMode != "" {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case "HTML":
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(text, mode)
		if err == nil && result != nil {
			text = result.Text
			ents = result.Entities
		}
	}

	req := &tg.MessagesSendMessageRequest{
		Peer:     peer,
		Message:  text,
		Entities: ents,
	}

	if markup != nil {
		req.ReplyMarkup = markup
	}
	if noWebpage {
		req.NoWebpage = true
	}
	if silent {
		req.Silent = true
	}

	upds, err := c.RawClient.MessagesSendMessage(c.Ctx, req)
	if err != nil {
		return nil, err
	}

	msg, err := functions.ReturnNewMessageWithError(&tg.Message{Message: text}, upds, c.PeerStorage, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(msg, c.Ctx, c.RawClient, c.PeerStorage, c.SelfID), nil
}

// SendMedia sends media to this channel.
// Opts can be:
//   - string (caption parse mode): "HTML", "Markdown", "MarkdownV2", or ""
//   - []tg.MessageEntityClass (raw caption entities)
//   - struct with fields: Caption, ParseMode, Markup, Silent
func (c *Channel) SendMedia(media tg.InputMediaClass, opts ...any) (*Message, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("channel has no client context")
	}
	peer := c.PeerStorage.GetInputPeerByID(c.GetID())
	if peer == nil {
		peer = &tg.InputPeerChannel{
			ChannelID:  c.ID,
			AccessHash: c.AccessHash,
		}
	}

	var caption string
	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var silent bool
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		if s, ok := opts[0].(string); ok {
			caption = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			ents = e
		} else {
			val := reflect.ValueOf(opts[0])
			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				for i := 0; i < val.NumField(); i++ {
					field := val.Type().Field(i)
					value := val.Field(i)
					if !value.IsValid() {
						continue
					}
					switch field.Name {
					case "Caption":
						if c, ok := value.Interface().(string); ok {
							caption = c
						}
					case "Entities":
						if e, ok := value.Interface().([]tg.MessageEntityClass); ok {
							ents = e
						}
					case "Markup":
						if rm, ok := value.Interface().(tg.ReplyMarkupClass); ok {
							markup = rm
						}
					case "Silent":
						if s, ok := value.Interface().(bool); ok {
							silent = s
						}
					case "ParseMode":
						if s, ok := value.Interface().(string); ok {
							parseMode = s
						}
					}
				}
			}
		}
	}

	if len(ents) == 0 && parseMode != "" {
		var mode parsemode.ParseMode
		switch strings.ToUpper(strings.TrimSpace(parseMode)) {
		case "HTML":
			mode = parsemode.ModeHTML
		case "MARKDOWN", "MARKDOWNV2":
			mode = parsemode.ModeMarkdown
		default:
			mode = parsemode.ModeNone
		}

		result, err := parsemode.Parse(caption, mode)
		if err == nil && result != nil {
			caption = result.Text
			ents = result.Entities
		}
	}

	req := &tg.MessagesSendMediaRequest{
		Peer:     peer,
		Media:    media,
		Message:  caption,
		Entities: ents,
	}

	if markup != nil {
		req.ReplyMarkup = markup
	}
	if silent {
		req.Silent = true
	}

	upds, err := c.RawClient.MessagesSendMedia(c.Ctx, req)
	if err != nil {
		return nil, err
	}

	msg, err := functions.ReturnNewMessageWithError(&tg.Message{Message: caption}, upds, c.PeerStorage, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(msg, c.Ctx, c.RawClient, c.PeerStorage, c.SelfID), nil
}
