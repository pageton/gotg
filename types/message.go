package types

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
	"github.com/pageton/gotg/storage"
)

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

func ConstructMessage(m tg.MessageClass) *Message {
	return ConstructMessageWithContext(m, nil, nil, nil, 0)
}

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

func (m *Message) SetRepliedToMessage(ctx context.Context, raw *tg.Client, p *storage.PeerStorage) error {
	replyMessage, ok := m.ReplyTo.(*tg.MessageReplyHeader)
	if !ok {
		return errors.ErrReplyNotMessage
	}
	replyTo := replyMessage.ReplyToMsgID
	if replyTo == 0 {
		return errors.ErrMessageNotExist
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
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
func (m *Message) Delete() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	_, err := m.RawClient.MessagesDeleteMessages(m.Ctx, &tg.MessagesDeleteMessagesRequest{
		ID:     []int{m.ID},
		Revoke: true,
	})
	return err
}

// Edit edits the message text.
// Opts can be:
//   - *adapter.ReplyOpts - from ext package
//   - []tg.MessageEntityClass - raw entities for backward compatibility
func (m *Message) Edit(text string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var noWebpage bool
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		// Handle []tg.MessageEntityClass for backward compatibility
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			ents = e
		} else {
			// Extract fields from struct (*adapter.ReplyOpts, etc.)
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
		switch parseMode {
		case "HTML":
			mode = parsemode.ModeHTML
		case "Markdown", "MarkdownV2":
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

	req := &tg.MessagesEditMessageRequest{
		Peer:     peer,
		ID:       m.ID,
		Message:  text,
		Entities: ents,
	}

	if markup != nil {
		req.ReplyMarkup = markup
	}
	if noWebpage {
		req.NoWebpage = true
	}

	upds, err := m.RawClient.MessagesEditMessage(m.Ctx, req)
	if err != nil {
		return nil, err
	}

	message, err := functions.ReturnEditMessageWithError(m.PeerStorage, upds, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(message, m.Ctx, m.RawClient, m.PeerStorage, m.SelfID), nil
}

// EditCaption edits the caption of a media message.
// Opts can be *adapter.ReplyOpts or []tg.MessageEntityClass for backward compatibility
func (m *Message) EditCaption(caption string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		// Handle []tg.MessageEntityClass for backward compatibility
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			ents = e
		} else {
			// Extract fields from struct (*adapter.ReplyOpts, etc.)
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
		switch parseMode {
		case "HTML":
			mode = parsemode.ModeHTML
		case "Markdown", "MarkdownV2":
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

	req := &tg.MessagesEditMessageRequest{
		Peer:     peer,
		ID:       m.ID,
		Message:  caption,
		Entities: ents,
	}

	if markup != nil {
		req.ReplyMarkup = markup
	}

	upds, err := m.RawClient.MessagesEditMessage(m.Ctx, req)
	if err != nil {
		return nil, err
	}

	message, err := functions.ReturnEditMessageWithError(m.PeerStorage, upds, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(message, m.Ctx, m.RawClient, m.PeerStorage, m.SelfID), nil
}

// EditReplyMarkup edits only the reply markup of the message.
func (m *Message) EditReplyMarkup(markup tg.ReplyMarkupClass) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	req := &tg.MessagesEditMessageRequest{
		Peer:        peer,
		ID:          m.ID,
		ReplyMarkup: markup,
	}

	upds, err := m.RawClient.MessagesEditMessage(m.Ctx, req)
	if err != nil {
		return nil, err
	}

	message, err := functions.ReturnEditMessageWithError(m.PeerStorage, upds, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(message, m.Ctx, m.RawClient, m.PeerStorage, m.SelfID), nil
}

// Reply sends a reply to this message.
// Opts can be *adapter.ReplyOpts or []tg.MessageEntityClass for backward compatibility
func (m *Message) Reply(text string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var noWebpage, silent bool

	if len(opts) > 0 && opts[0] != nil {
		// Handle []tg.MessageEntityClass for backward compatibility
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			ents = e
		} else {
			// Extract fields from struct (*adapter.ReplyOpts, etc.)
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
					case "WithoutReply":
						if wr, ok := value.Interface().(bool); ok {
							if wr {
								peer = nil
							}
						}
					}
				}
			}
		}
	}

	req := &tg.MessagesSendMessageRequest{
		Peer:     peer,
		Message:  text,
		Entities: ents,
		ReplyTo: &tg.InputReplyToMessage{
			ReplyToMsgID: m.ID,
		},
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

	upds, err := m.RawClient.MessagesSendMessage(m.Ctx, req)
	if err != nil {
		return nil, err
	}

	msg, err := functions.ReturnNewMessageWithError(&tg.Message{Message: text}, upds, m.PeerStorage, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(msg, m.Ctx, m.RawClient, m.PeerStorage, m.SelfID), nil
}

// ReplyMedia sends a media reply to this message.
// Opts can be *adapter.ReplyMediaOpts or []tg.MessageEntityClass for backward compatibility
func (m *Message) ReplyMedia(media tg.InputMediaClass, caption string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var silent bool

	if len(opts) > 0 && opts[0] != nil {
		// Handle []tg.MessageEntityClass for backward compatibility
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
			ents = e
		} else {
			// Extract fields from struct (*adapter.ReplyMediaOpts, etc.)
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
					case "Silent":
						if s, ok := value.Interface().(bool); ok {
							silent = s
						}
					case "WithoutReply":
						if wr, ok := value.Interface().(bool); ok {
							if wr {
								peer = nil
							}
						}
					}
				}
			}
		}
	}

	req := &tg.MessagesSendMediaRequest{
		Peer:     peer,
		Media:    media,
		Message:  caption,
		Entities: ents,
		ReplyTo: &tg.InputReplyToMessage{
			ReplyToMsgID: m.ID,
		},
	}

	if markup != nil {
		req.ReplyMarkup = markup
	}
	if silent {
		req.Silent = true
	}

	upds, err := m.RawClient.MessagesSendMedia(m.Ctx, req)
	if err != nil {
		return nil, err
	}

	msg, err := functions.ReturnNewMessageWithError(&tg.Message{Message: caption}, upds, m.PeerStorage, err)
	if err != nil {
		return nil, err
	}

	return ConstructMessageWithContext(msg, m.Ctx, m.RawClient, m.PeerStorage, m.SelfID), nil
}

// Pin pins this message in the chat.
func (m *Message) Pin() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)
	_, err := m.RawClient.MessagesUpdatePinnedMessage(m.Ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer: peer,
		ID:   m.ID,
	})
	return err
}

// Unpin unpins this message from the chat.
func (m *Message) Unpin() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)
	_, err := m.RawClient.MessagesUpdatePinnedMessage(m.Ctx, &tg.MessagesUpdatePinnedMessageRequest{
		Peer:  peer,
		ID:    m.ID,
		Unpin: true,
	})
	return err
}

// UnpinAllMessages unpins all messages in this message's chat.
func (m *Message) UnpinAllMessages() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIdFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)
	_, err := m.RawClient.MessagesUnpinAllMessages(m.Ctx, &tg.MessagesUnpinAllMessagesRequest{
		Peer: peer,
	})
	return err
}

// Download downloads the media from this message to a file path.
// If path is empty, auto-generates a filename using GetMediaFileNameWithID.
// Returns the file path that was used and the file type.
//
// Example:
//
//	path, fileType, err := msg.Download("")
//	path, fileType, err := msg.Download("downloads/photo.jpg")
func (m *Message) Download(path string) (string, tg.StorageFileTypeClass, error) {
	if m.RawClient == nil {
		return "", nil, fmt.Errorf("message has no client context")
	}
	if m.Media == nil {
		return "", nil, fmt.Errorf("message has no media")
	}

	// Auto-generate filename if path is empty
	if path == "" {
		var err error
		path, err = functions.GetMediaFileNameWithID(m.Media)
		if err != nil {
			return "", nil, err
		}
	}

	inputFileLocation, err := functions.GetInputFileLocation(m.Media)
	if err != nil {
		return "", nil, err
	}

	mediaDownloader := downloader.NewDownloader()
	d := mediaDownloader.Download(m.RawClient, inputFileLocation)

	fileType, err := d.ToPath(m.Ctx, path)
	if err != nil {
		return "", nil, err
	}

	return path, fileType, nil
}

// DownloadBytes downloads the media from this message into memory.
// Returns the file content as bytes and the file type.
//
// Example:
//
//	data, fileType, err := msg.DownloadBytes()
func (m *Message) DownloadBytes() ([]byte, tg.StorageFileTypeClass, error) {
	if m.RawClient == nil {
		return nil, nil, fmt.Errorf("message has no client context")
	}
	if m.Media == nil {
		return nil, nil, fmt.Errorf("message has no media")
	}

	inputFileLocation, err := functions.GetInputFileLocation(m.Media)
	if err != nil {
		return nil, nil, err
	}

	mediaDownloader := downloader.NewDownloader()
	d := mediaDownloader.Download(m.RawClient, inputFileLocation)

	// Use a buffer writer to collect bytes
	var buf []byte
	fileType, err := d.Stream(m.Ctx, &byteWriter{b: &buf})
	if err != nil {
		return nil, nil, err
	}

	return buf, fileType, nil
}

// byteWriter helps collect bytes into a slice.
type byteWriter struct {
	b *[]byte
}

func (bw *byteWriter) Write(p []byte) (n int, err error) {
	*bw.b = append(*bw.b, p...)
	return len(p), nil
}

// GetUser fetches and returns the full user info of the message sender.
// Returns nil if the message has no sender or on error.
//
// Example:
//
//	// Get sender of effective message
//	user, err := u.EffectiveMessage.GetUser()
//
//	// Get sender of replied message
//	user, err := u.EffectiveReply().GetUser()
func (m *Message) GetUser() (*tg.UserFull, error) {
	if m == nil {
		return nil, fmt.Errorf("message is nil")
	}
	if m.Message == nil {
		return nil, fmt.Errorf("message is nil")
	}
	if m.FromID == nil {
		return nil, fmt.Errorf("message has no sender")
	}

	peerUser, ok := m.FromID.(*tg.PeerUser)
	if !ok {
		return nil, fmt.Errorf("sender is not a user")
	}

	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}

	user, err := m.RawClient.UsersGetFullUser(m.Ctx, &tg.InputUser{
		UserID:     peerUser.UserID,
		AccessHash: 0,
	})
	if err != nil {
		return nil, err
	}

	return &user.FullUser, nil
}
