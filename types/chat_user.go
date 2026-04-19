package types

import (
	"context"
	"fmt"
	"html"
	"reflect"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/storage"
)

// User implements EffectiveChat interface for tg.User chats.
type User struct {
	tg.User
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// GetID returns the user ID.
func (u *User) GetID() int64 {
	return u.ID
}

// GetAccessHash returns the access hash for API authentication.
func (u *User) GetAccessHash() int64 {
	return u.AccessHash
}

// GetInputUser returns the InputUser for API calls.
func (v *User) GetInputUser() tg.InputUserClass {
	return &tg.InputUser{
		UserID:     v.ID,
		AccessHash: v.AccessHash,
	}
}

// GetInputChannel returns the InputChannel for this peer.
// Always returns nil for User (users are not channels).
func (*User) GetInputChannel() tg.InputChannelClass {
	return nil
}

// GetInputPeer returns the InputPeer for API calls.
func (v *User) GetInputPeer() tg.InputPeerClass {
	return &tg.InputPeerUser{
		UserID:     v.ID,
		AccessHash: v.AccessHash,
	}
}

// IsChannel returns true if this effective chat is a channel.
// Always returns false for User.
func (*User) IsChannel() bool {
	return false
}

// IsChat returns true if this effective chat is a group chat.
// Always returns false for User.
func (*User) IsChat() bool {
	return false
}

// IsUser returns true if this effective chat is a user (private chat).
// Always returns true for User.
func (*User) IsUser() bool {
	return true
}

// Raw returns the underlying tg.User struct.
func (u *User) Raw() *tg.User {
	return &u.User
}

// GetChat returns the basic chat information for this user.
// Returns tg.ChatClass which can be type-asserted to *tg.User.
func (u *User) GetChat() (tg.ChatClass, error) {
	if u.RawClient == nil {
		return nil, fmt.Errorf("user has no client context")
	}
	return functions.GetChat(u.Ctx, u.RawClient, u.PeerStorage, u.ID)
}

// Mention generates an HTML mention link for this user.
//
// Behavior:
// - No arguments: uses the user's ID and full name.
// - One argument:
//   - int/int64 → overrides userID, keeps default name.
//   - string → overrides name, keeps default userID.
//
// - Two arguments: first is userID (int/int64), second is name (string).
// - Returns a string in the format: <a href='tg://user?id=USERID'>NAME</a>
func (u *User) Mention(args ...any) string {
	userID := u.ID
	name := u.FirstName + " " + u.LastName

	if len(args) == 1 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		case string:
			name = html.EscapeString(v)
		}
	} else if len(args) >= 2 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		}
		if n, ok := args[1].(string); ok {
			name = html.EscapeString(n)
		}
	}

	return fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", userID, name)
}

// SendMessage sends a message to this user.
// Opts can be:
//   - string (parse mode): "HTML", "Markdown", "MarkdownV2", or ""
//   - []tg.MessageEntityClass (raw entities for backward compatibility)
//   - struct with fields: ParseMode, NoWebpage, Silent, Markup
func (u *User) SendMessage(text string, opts ...any) (*Message, error) {
	if u.RawClient == nil {
		return nil, fmt.Errorf("user has no client context")
	}
	peer := u.PeerStorage.GetInputPeerByID(u.ID)
	if peer == nil {
		peer = &tg.InputPeerUser{
			UserID:     u.ID,
			AccessHash: u.AccessHash,
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

	upds, err := u.RawClient.MessagesSendMessage(u.Ctx, req)
	if err != nil {
		return nil, err
	}

	msg, err := functions.ReturnNewMessageWithError(&tg.Message{Message: text}, upds, u.PeerStorage, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(msg, u.Ctx, u.RawClient, u.PeerStorage, u.SelfID), nil
}

// SendMedia sends media to this user.
// Opts can be:
//   - string (caption parse mode): "HTML", "Markdown", "MarkdownV2", or ""
//   - []tg.MessageEntityClass (raw caption entities)
//   - struct with fields: Caption, ParseMode, Markup, Silent
func (u *User) SendMedia(media tg.InputMediaClass, opts ...any) (*Message, error) {
	if u.RawClient == nil {
		return nil, fmt.Errorf("user has no client context")
	}
	peer := u.PeerStorage.GetInputPeerByID(u.ID)
	if peer == nil {
		peer = &tg.InputPeerUser{
			UserID:     u.ID,
			AccessHash: u.AccessHash,
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

	upds, err := u.RawClient.MessagesSendMedia(u.Ctx, req)
	if err != nil {
		return nil, err
	}

	msg, err := functions.ReturnNewMessageWithError(&tg.Message{Message: caption}, upds, u.PeerStorage, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(msg, u.Ctx, u.RawClient, u.PeerStorage, u.SelfID), nil
}

// Block blocks this user.
func (u *User) Block() error {
	if u.RawClient == nil {
		return fmt.Errorf("user has no client context")
	}
	_, err := u.RawClient.ContactsBlock(u.Ctx, &tg.ContactsBlockRequest{
		ID: &tg.InputPeerUser{
			UserID:     u.ID,
			AccessHash: u.AccessHash,
		},
	})
	return err
}

// Unblock unblocks this user.
func (u *User) Unblock() error {
	if u.RawClient == nil {
		return fmt.Errorf("user has no client context")
	}
	_, err := u.RawClient.ContactsUnblock(u.Ctx, &tg.ContactsUnblockRequest{
		ID: &tg.InputPeerUser{
			UserID:     u.ID,
			AccessHash: u.AccessHash,
		},
	})
	return err
}

// GetChatMemberIn fetches this user's membership info in a specific chat.
// chatID can be a channel, chat, or user ID.
func (u *User) GetChatMemberIn(chatID int64) (any, error) {
	if u.RawClient == nil {
		return nil, fmt.Errorf("user has no client context")
	}

	peer := u.PeerStorage.GetInputPeerByID(chatID)
	if peer == nil {
		return nil, fmt.Errorf("chat not found in peer storage")
	}

	userPeer := &tg.InputPeerUser{
		UserID:     u.ID,
		AccessHash: u.AccessHash,
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		res, err := u.RawClient.ChannelsGetParticipant(u.Ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  p.ChannelID,
				AccessHash: p.AccessHash,
			},
			Participant: userPeer,
		})
		if err != nil {
			return nil, err
		}
		return res.Participant, nil

	case *tg.InputPeerChat:
		fullChat, err := u.RawClient.MessagesGetFullChat(u.Ctx, p.ChatID)
		if err != nil {
			return nil, err
		}

		if fc, ok := fullChat.FullChat.(*tg.ChatFull); ok {
			if fc.Participants == nil {
				return nil, fmt.Errorf("chat participants not available")
			}

			if pt, ok := fc.Participants.(*tg.ChatParticipants); ok {
				for _, participant := range pt.Participants {
					switch cp := participant.(type) {
					case *tg.ChatParticipant:
						if cp.UserID == u.ID {
							return cp, nil
						}
					case *tg.ChatParticipantCreator:
						if cp.UserID == u.ID {
							return cp, nil
						}
					case *tg.ChatParticipantAdmin:
						if cp.UserID == u.ID {
							return cp, nil
						}
					}
				}
				return nil, fmt.Errorf("user not found in chat")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported peer type for GetChatMember")
	}

	return nil, fmt.Errorf("failed to get chat member")
}
