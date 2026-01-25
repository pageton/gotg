package types

import (
	"context"
	"fmt"
	"html"
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
	IsChannel() bool
	// Use this method to check if the effective chat is a chat.
	IsChat() bool
	// Use this method to check if the effective chat is a user.
	IsUser() bool
	// Use this method to get InputUserClass
	GetInputUser() tg.InputUserClass
	// Use this method to get InputUserClass
	GetInputChannel() tg.InputChannelClass
	// Use this method to get InputUserClass
	GetInputPeer() tg.InputPeerClass
}

// Participant represents a chat/channel member with their details.
type Participant struct {
	// The user object containing user information
	User *User
	// The raw channel participant (for channels, nil for other types)
	Participant tg.ChannelParticipantClass
	// The member status (Creator, Admin, Member, Restricted, Left, etc.)
	Status string
	// The admin rights (for admins/creators)
	Rights *tg.ChatAdminRights
	// The custom rank/title (for admins/creators)
	Title string
	// The user ID
	UserID int64
	// The chat ID
	ChatID int64
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

// User implements EffectiveChat interface for tg.User chats.
type User struct {
	tg.User
	// Context fields for bound methods
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
