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
	// Use this method to get chat id.
	GetID() int64
	// Use this method to get access hash of the effective chat.
	GetAccessHash() int64
	// Use this method to check if the effective chat is a channel.
	IsAChannel() bool
	// Use this method to check if the effective chat is a chat.
	IsAChat() bool
	// Use this method to check if the effective chat is a user.
	IsAUser() bool
	// Use this method to get InputUserClass
	GetInputUser() tg.InputUserClass
	// Use this method to get InputUserClass
	GetInputChannel() tg.InputChannelClass
	// Use this method to get InputUserClass
	GetInputPeer() tg.InputPeerClass
}

// EmptyUC implements EffectiveChat interface for empty chats.
type EmptyUC struct{}

// Use this method to get chat id.
// Always 0 for EmptyUC
func (*EmptyUC) GetID() int64 {
	return 0
}

// Use this method to get access hash of effective chat.
// Always 0 for EmptyUC
func (*EmptyUC) GetAccessHash() int64 {
	return 0
}

// Use this method to get InputUserClass
// Always nil for EmptyUC
func (*EmptyUC) GetInputUser() tg.InputUserClass {
	return nil
}

// Use this method to get InputChannelClass
// Always nil for EmptyUC
func (*EmptyUC) GetInputChannel() tg.InputChannelClass {
	return nil
}

// Use this method to get InputPeerClass
// Always nil for EmptyUC
func (*EmptyUC) GetInputPeer() tg.InputPeerClass {
	return nil
}

// IsAChannel returns true for a channel.
// Always false for EmptyUC
func (*EmptyUC) IsAChannel() bool {
	return false
}

// IsAChat returns true for a chat.
// Always false for EmptyUC
func (*EmptyUC) IsAChat() bool {
	return false
}

// IsAUser returns true for a user.
// Always false for EmptyUC
func (*EmptyUC) IsAUser() bool {
	return false
}

// User implements EffectiveChat interface for tg.User chats.
type User struct {
	tg.User
	// Context fields for bound methods
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// Use this method to get chat id.
func (u *User) GetID() int64 {
	return u.ID
}

// Use this method to get access hash of the effective chat.
func (u *User) GetAccessHash() int64 {
	return u.AccessHash
}

// Use this method to get InputUserClass
func (v *User) GetInputUser() tg.InputUserClass {
	return &tg.InputUser{
		UserID:     v.ID,
		AccessHash: v.AccessHash,
	}
}

// Use this method to get InputChannelClass
// Always nil for User
func (*User) GetInputChannel() tg.InputChannelClass {
	return nil
}

// Use this method to get InputPeerClass
func (v *User) GetInputPeer() tg.InputPeerClass {
	return &tg.InputPeerUser{
		UserID:     v.ID,
		AccessHash: v.AccessHash,
	}
}

// IsAChannel returns true for a channel.
func (*User) IsAChannel() bool {
	return false
}

// IsAChat returns true for a chat.
func (*User) IsAChat() bool {
	return false
}

// IsAUser returns true for a user.
func (*User) IsAUser() bool {
	return true
}

func (u *User) Raw() *tg.User {
	return &u.User
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
			name = v
		}
	} else if len(args) >= 2 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		}
		if n, ok := args[1].(string); ok {
			name = n
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
		// Handle string as parse mode
		if s, ok := opts[0].(string); ok {
			parseMode = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			// Handle []tg.MessageEntityClass for backward compatibility
			ents = e
		} else {
			// Extract fields from struct
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

	// Parse HTML/Markdown to entities if parseMode is set and ents is not provided
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
		// Handle string as caption
		if s, ok := opts[0].(string); ok {
			caption = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			// Handle []tg.MessageEntityClass for caption entities
			ents = e
		} else {
			// Extract fields from struct
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

	// Parse HTML/Markdown to entities if parseMode is set and ents is not provided
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

	// Create the input peer for this user
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

		switch fc := fullChat.FullChat.(type) {
		case *tg.ChatFull:
			if fc.Participants == nil {
				return nil, fmt.Errorf("chat participants not available")
			}

			switch pt := fc.Participants.(type) {
			case *tg.ChatParticipants:
				for _, participant := range pt.Participants {
					switch cp := participant.(type) {
					case *tg.ChatParticipant, *tg.ChatParticipantCreator, *tg.ChatParticipantAdmin:
						if cp.(*tg.ChatParticipant).UserID == u.ID {
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

// Channel implements EffectiveChat interface for tg.Channel chats.
type Channel struct {
	tg.Channel
	// Context fields for bound methods
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// Use this method to get chat id.
func (u *Channel) GetID() int64 {
	var ID constant.TDLibPeerID
	ID.Channel(u.ID)
	return int64(ID)
}

// Use this method to get access hash of the effective chat.
func (u *Channel) GetAccessHash() int64 {
	return u.AccessHash
}

// Use this method to get InputUserClass
// Always nil for Channel
func (*Channel) GetInputUser() tg.InputUserClass {
	return nil
}

// Use this method to get InputChannelClass
func (v *Channel) GetInputChannel() tg.InputChannelClass {
	return &tg.InputChannel{
		ChannelID:  v.ID,
		AccessHash: v.AccessHash,
	}
}

// Use this method to get InputPeerClass
func (v *Channel) GetInputPeer() tg.InputPeerClass {
	return &tg.InputPeerChannel{
		ChannelID:  v.ID,
		AccessHash: v.AccessHash,
	}
}

// IsAChannel returns true for a channel.
func (*Channel) IsAChannel() bool {
	return true
}

// IsAChat returns true for a chat.
func (*Channel) IsAChat() bool {
	return false
}

// IsAUser returns true for a user.
func (*Channel) IsAUser() bool {
	return false
}

func (c *Channel) Raw() *tg.Channel {
	return &c.Channel
}

// GetChatMember fetches information about a specific member of this channel.
// Returns tg.ChannelParticipantClass with the member information.
func (c *Channel) GetChatMember(userID int64) (tg.ChannelParticipantClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("channel has no client context")
	}

	// Get the input peer for the user
	peer := c.PeerStorage.GetInputPeerByID(userID)
	if peer == nil {
		// If not in peer storage, create a basic InputPeerUser
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
		// Handle string as parse mode
		if s, ok := opts[0].(string); ok {
			parseMode = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			// Handle []tg.MessageEntityClass for backward compatibility
			ents = e
		} else {
			// Extract fields from struct
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

	// Parse HTML/Markdown to entities if parseMode is set and ents is not provided
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
		// Handle string as caption
		if s, ok := opts[0].(string); ok {
			caption = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			// Handle []tg.MessageEntityClass for caption entities
			ents = e
		} else {
			// Extract fields from struct
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

	// Parse HTML/Markdown to entities if parseMode is set and ents is not provided
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

// Chat implements EffectiveChat interface for tg.Chat chats.
type Chat struct {
	tg.Chat
	// Context fields for bound methods
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
}

// Use this method to get chat id.
func (u *Chat) GetID() int64 {
	var ID constant.TDLibPeerID
	ID.Chat(u.ID)
	return int64(ID)
}

// Use this method to get access hash of the effective chat.
func (*Chat) GetAccessHash() int64 {
	return 0
}

// Use this method to get InputUserClass
// Always nil for Chat
func (*Chat) GetInputUser() tg.InputUserClass {
	return nil
}

// Use this method to get InputChannelClass
// Always nil for Chat
func (*Chat) GetInputChannel() tg.InputChannelClass {
	return nil
}

// Use this method to get InputPeerClass
func (v *Chat) GetInputPeer() tg.InputPeerClass {
	return &tg.InputPeerChat{
		ChatID: v.ID,
	}
}

// IsAChannel returns true for a channel.
func (*Chat) IsAChannel() bool {
	return false
}

// IsAChat returns true for a chat.
func (*Chat) IsAChat() bool {
	return true
}

// IsAUser returns true for a user.
func (*Chat) IsAUser() bool {
	return false
}

func (c *Chat) Raw() *tg.Chat {
	return &c.Chat
}

// GetChatMember fetches information about a specific member of this chat.
// For normal chats, this returns the chat participant info.
func (c *Chat) GetChatMember(userID int64) (tg.ChatParticipantClass, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("chat has no client context")
	}

	// Get full chat info which includes participants
	fullChat, err := c.RawClient.MessagesGetFullChat(c.Ctx, c.ID)
	if err != nil {
		return nil, err
	}

	// Extract the participants from the full chat
	switch fc := fullChat.FullChat.(type) {
	case *tg.ChatFull:
		if fc.Participants == nil {
			return nil, fmt.Errorf("chat participants not available")
		}

		// Search through participants to find the matching user
		switch p := fc.Participants.(type) {
		case *tg.ChatParticipants:
			for _, participant := range p.Participants {
				switch cp := participant.(type) {
				case *tg.ChatParticipant:
					if cp.UserID == userID {
						return cp, nil
					}
				case *tg.ChatParticipantCreator:
					if cp.UserID == userID {
						return cp, nil
					}
				case *tg.ChatParticipantAdmin:
					if cp.UserID == userID {
						return cp, nil
					}
				}
			}
			return nil, fmt.Errorf("user not found in chat")
		default:
			return nil, fmt.Errorf("unexpected participants type")
		}
	default:
		return nil, fmt.Errorf("unexpected full chat type")
	}
}

// SendMessage sends a message to this chat.
// Opts can be:
//   - string (parse mode): "HTML", "Markdown", "MarkdownV2", or ""
//   - []tg.MessageEntityClass (raw entities for backward compatibility)
//   - struct with fields: ParseMode, NoWebpage, Silent, Markup
func (c *Chat) SendMessage(text string, opts ...any) (*Message, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("chat has no client context")
	}
	peer := c.PeerStorage.GetInputPeerByID(c.GetID())
	if peer == nil {
		peer = &tg.InputPeerChat{
			ChatID: c.ID,
		}
	}

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var noWebpage, silent bool
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		// Handle string as parse mode
		if s, ok := opts[0].(string); ok {
			parseMode = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			// Handle []tg.MessageEntityClass for backward compatibility
			ents = e
		} else {
			// Extract fields from struct
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

	// Parse HTML/Markdown to entities if parseMode is set and ents is not provided
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

// SendMedia sends media to this chat.
// Opts can be:
//   - string (caption parse mode): "HTML", "Markdown", "MarkdownV2", or ""
//   - []tg.MessageEntityClass (raw caption entities)
//   - struct with fields: Caption, ParseMode, Markup, Silent
func (c *Chat) SendMedia(media tg.InputMediaClass, opts ...any) (*Message, error) {
	if c.RawClient == nil {
		return nil, fmt.Errorf("chat has no client context")
	}
	peer := c.PeerStorage.GetInputPeerByID(c.GetID())
	if peer == nil {
		peer = &tg.InputPeerChat{
			ChatID: c.ID,
		}
	}

	var caption string
	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var silent bool
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		// Handle string as caption
		if s, ok := opts[0].(string); ok {
			caption = s
		} else if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			// Handle []tg.MessageEntityClass for caption entities
			ents = e
		} else {
			// Extract fields from struct
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

	// Parse HTML/Markdown to entities if parseMode is set and ents is not provided
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
