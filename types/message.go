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

// Edit edits the message text.
// Opts can be:
//   - *adapter.ReplyOpts - from ext package
//   - []tg.MessageEntityClass - raw entities for backward compatibility
func (m *Message) Edit(text string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
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
	chatID := functions.GetChatIDFromPeer(m.PeerID)
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
	chatID := functions.GetChatIDFromPeer(m.PeerID)
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
//
// NOTE: By default, replies to this message. To reply to a different message,
// use ReplyOpts with ReplyMessageID set.
func (m *Message) Reply(text string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var noWebpage, silent bool
	var replyMessageID int
	var withoutReply bool

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
							withoutReply = wr
						}
					case "ReplyMessageID":
						if rid, ok := value.Interface().(int); ok {
							replyMessageID = rid
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
		RandomID: functions.GenerateRandomID(),
	}

	// Set reply to - either custom message ID, this message, or no reply
	if replyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: replyMessageID,
		}
	} else if !withoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: m.ID,
		}
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
// Accepts tg.InputMediaClass (e.g., &tg.InputMediaPhoto{}, &tg.InputMediaDocument{}).
// Opts can be *adapter.ReplyMediaOpts or []tg.MessageEntityClass for backward compatibility
//
// NOTE: By default, replies to this message. To reply to a different message,
// use ReplyMediaOpts with ReplyMessageID set.
//
// Example using InputMedia:
//
//	newMsg, err := msg.ReplyMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, "Photo caption")
//
// For using fileID strings, see ReplyMediaWithFileID.
func (m *Message) ReplyMedia(media tg.InputMediaClass, caption string, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var silent bool
	var replyMessageID int
	var withoutReply bool

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
							withoutReply = wr
						}
					case "ReplyMessageID":
						if rid, ok := value.Interface().(int); ok {
							replyMessageID = rid
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
	}

	// Set reply to - either custom message ID, this message, or no reply
	if replyMessageID != 0 {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: replyMessageID,
		}
	} else if !withoutReply {
		req.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: m.ID,
		}
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

// ReplyMediaWithFileID sends a media reply to this message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
// Opts can be *adapter.ReplyMediaOpts or []tg.MessageEntityClass for backward compatibility
//
// Example:
//
//	// Get fileID from a photo message
//	fileID := photoMsg.FileID()
//
//	// Reply to another message with the same media
//	newMsg, err := originalMsg.ReplyMediaWithFileID(fileID, "Here's the media you requested")
func (m *Message) ReplyMediaWithFileID(fileID string, caption string, opts ...any) (*Message, error) {
	media, err := InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return m.ReplyMedia(media, caption, opts...)
}

// EditMedia edits the media of this message.
// Accepts tg.InputMediaClass (e.g., &tg.InputMediaPhoto{}, &tg.InputMediaDocument{}).
// Opts can be *adapter.ReplyMediaOpts for caption and entities.
//
// Example using InputMedia:
//
//	editedMsg, err := msg.EditMedia(&tg.InputMediaPhoto{
//	    ID: &tg.InputPhoto{ID: photoID, AccessHash: accessHash},
//	}, &adapter.ReplyMediaOpts{
//	    Caption: "New photo",
//	})
//
// For using fileID strings, see EditMediaWithFileID.
func (m *Message) EditMedia(media tg.InputMediaClass, opts ...any) (*Message, error) {
	if m.RawClient == nil {
		return nil, fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	peer := m.PeerStorage.GetInputPeerByID(chatID)

	// Extract opts for caption and entities
	var caption string
	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var parseMode string

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
					case "Caption":
						if s, ok := value.Interface().(string); ok {
							caption = s
						}
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
		Media:    media,
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

// EditMediaWithFileID edits the media of this message using a fileID string.
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
// Opts can be *adapter.ReplyMediaOpts for caption and entities.
//
// Example:
//
//	// Replace message media with a fileID
//	editedMsg, err := msg.EditMediaWithFileID(newFileID, &adapter.ReplyMediaOpts{
//	    Caption: "Updated media!",
//	})
func (m *Message) EditMediaWithFileID(fileID string, opts ...any) (*Message, error) {
	// Extract caption from opts to pass to InputMediaFromFileID
	var caption string
	if len(opts) > 0 && opts[0] != nil {
		val := reflect.ValueOf(opts[0])
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		}
		if val.Kind() == reflect.Struct {
			for i := 0; i < val.NumField(); i++ {
				if val.Type().Field(i).Name == "Caption" {
					if s, ok := val.Field(i).Interface().(string); ok {
						caption = s
					}
					break
				}
			}
		}
	}

	media, err := InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return m.EditMedia(media, opts...)
}

// Pin pins this message in the chat.
func (m *Message) Pin() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	_, err := functions.PinMessage(m.Ctx, m.RawClient, m.PeerStorage, chatID, m.ID)
	return err
}

// Unpin unpins this message from the chat.
func (m *Message) Unpin() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	return functions.UnpinMessage(m.Ctx, m.RawClient, m.PeerStorage, chatID, m.ID)
}

// UnpinAllMessages unpins all messages in this message's chat.
func (m *Message) UnpinAllMessages() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	return functions.UnpinAllMessages(m.Ctx, m.RawClient, m.PeerStorage, chatID)
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

	return functions.GetUser(m.Ctx, m.RawClient, m.PeerStorage, peerUser.UserID)
}

// Photo returns the photo if message contains photo media, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if photo := m.Photo(); photo != nil {
//		fmt.Printf("Photo ID: %d\n", photo.ID)
//		fmt.Printf("File ID: %s\n", photo.FileID())
//	}
func (m *Message) Photo() *Photo {
	if m == nil || m.Message == nil || m.Media == nil {
		return nil
	}
	if media, ok := m.Media.(*tg.MessageMediaPhoto); ok {
		if photo, ok := media.Photo.AsNotEmpty(); ok {
			return NewPhoto(photo)
		}
	}
	return nil
}

// Document returns the document if message contains document media, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.Document(); doc != nil {
//		fmt.Printf("Document name: %s\n", doc.GetAttribute().(*tg.DocumentAttributeFilename).FileName)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) Document() *Document {
	if m == nil || m.Message == nil || m.Media == nil {
		return nil
	}
	if media, ok := m.Media.(*tg.MessageMediaDocument); ok {
		if doc, ok := media.Document.(*tg.Document); ok {
			return NewDocument(doc)
		}
	}
	return nil
}

// Geo returns the location if message contains geo media, nil otherwise.
//
// Example:
//
//	if geo := m.Geo(); geo != nil {
//		fmt.Printf("Location: %f, %f\n", geo.Geo.Lat, geo.Geo.Long)
//	}
func (m *Message) Geo() *tg.MessageMediaGeo {
	if m.Media == nil {
		return nil
	}
	if geo, ok := m.Media.(*tg.MessageMediaGeo); ok {
		return geo
	}
	return nil
}

// GeoLive returns the live location if message contains geo live media, nil otherwise.
//
// Example:
//
//	if live := m.GeoLive(); live != nil {
//		fmt.Printf("Live location expires in: %d\n", live.Period)
//	}
func (m *Message) GeoLive() *tg.MessageMediaGeoLive {
	if m.Media == nil {
		return nil
	}
	if geoLive, ok := m.Media.(*tg.MessageMediaGeoLive); ok {
		return geoLive
	}
	return nil
}

// Contact returns the contact if message contains contact media, nil otherwise.
//
// Example:
//
//	if contact := m.Contact(); contact != nil {
//		fmt.Printf("Contact: %s (%s)\n", contact.FirstName, contact.PhoneNumber)
//	}
func (m *Message) Contact() *tg.MessageMediaContact {
	if m.Media == nil {
		return nil
	}
	if contact, ok := m.Media.(*tg.MessageMediaContact); ok {
		return contact
	}
	return nil
}

// Poll returns the poll if message contains poll media, nil otherwise.
//
// Example:
//
//	if poll := m.Poll(); poll != nil {
//		fmt.Printf("Poll question: %s\n", poll.Poll.Question)
//	}
func (m *Message) Poll() *tg.MessageMediaPoll {
	if m.Media == nil {
		return nil
	}
	if poll, ok := m.Media.(*tg.MessageMediaPoll); ok {
		return poll
	}
	return nil
}

// Venue returns the venue if message contains venue media, nil otherwise.
//
// Example:
//
//	if venue := m.Venue(); venue != nil {
//		fmt.Printf("Venue: %s\n", venue.Title)
//	}
func (m *Message) Venue() *tg.MessageMediaVenue {
	if m.Media == nil {
		return nil
	}
	if venue, ok := m.Media.(*tg.MessageMediaVenue); ok {
		return venue
	}
	return nil
}

// Dice returns the dice if message contains dice media, nil otherwise.
//
// Example:
//
//	if dice := m.Dice(); dice != nil {
//		fmt.Printf("Dice value: %d\n", dice.Value)
//	}
func (m *Message) Dice() *tg.MessageMediaDice {
	if m.Media == nil {
		return nil
	}
	if dice, ok := m.Media.(*tg.MessageMediaDice); ok {
		return dice
	}
	return nil
}

// Game returns the game if message contains game media, nil otherwise.
//
// Example:
//
//	if game := m.Game(); game != nil {
//		fmt.Printf("Game: %s\n", game.Game.Title)
//	}
func (m *Message) Game() *tg.MessageMediaGame {
	if m.Media == nil {
		return nil
	}
	if game, ok := m.Media.(*tg.MessageMediaGame); ok {
		return game
	}
	return nil
}

// WebPage returns the webpage if message contains webpage media, nil otherwise.
//
// Example:
//
//	if wp := m.WebPage(); wp != nil {
//		if webpage, ok := wp.Webpage.(*tg.WebPage); ok {
//			fmt.Printf("WebPage: %s\n", webpage.URL)
//		}
//	}
func (m *Message) WebPage() *tg.MessageMediaWebPage {
	if m.Media == nil {
		return nil
	}
	if wp, ok := m.Media.(*tg.MessageMediaWebPage); ok {
		return wp
	}
	return nil
}

// Invoice returns the invoice if message contains invoice media, nil otherwise.
//
// Example:
//
//	if invoice := m.Invoice(); invoice != nil {
//		fmt.Printf("Invoice: %s - %d\n", invoice.Title, invoice.TotalAmount)
//	}
func (m *Message) Invoice() *tg.MessageMediaInvoice {
	if m.Media == nil {
		return nil
	}
	if invoice, ok := m.Media.(*tg.MessageMediaInvoice); ok {
		return invoice
	}
	return nil
}

// Giveaway returns the giveaway if message contains giveaway media, nil otherwise.
//
// Example:
//
//	if giveaway := m.Giveaway(); giveaway != nil {
//		fmt.Printf("Giveaway prizes: %d\n", giveaway.Prizes)
//	}
func (m *Message) Giveaway() *tg.MessageMediaGiveaway {
	if m.Media == nil {
		return nil
	}
	if giveaway, ok := m.Media.(*tg.MessageMediaGiveaway); ok {
		return giveaway
	}
	return nil
}

// GiveawayResults returns the giveaway results if message contains giveaway results media, nil otherwise.
//
// Example:
//
//	if results := m.GiveawayResults(); results != nil {
//		fmt.Printf("Winners: %d\n", results.WinnersCount)
//	}
func (m *Message) GiveawayResults() *tg.MessageMediaGiveawayResults {
	if m.Media == nil {
		return nil
	}
	if results, ok := m.Media.(*tg.MessageMediaGiveawayResults); ok {
		return results
	}
	return nil
}

// Story returns the story if message contains story media, nil otherwise.
//
// Example:
//
//	if story := m.Story(); story != nil {
//		fmt.Printf("Story from peer: %d\n", story.Peer)
//	}
func (m *Message) Story() *tg.MessageMediaStory {
	if m.Media == nil {
		return nil
	}
	if story, ok := m.Media.(*tg.MessageMediaStory); ok {
		return story
	}
	return nil
}

// PaidMedia returns the paid media if message contains paid media, nil otherwise.
//
// Example:
//
//	if paid := m.PaidMedia(); paid != nil {
//		fmt.Printf("Stars amount: %d\n", paid.StarsAmount)
//	}
func (m *Message) PaidMedia() *tg.MessageMediaPaidMedia {
	if m.Media == nil {
		return nil
	}
	if paid, ok := m.Media.(*tg.MessageMediaPaidMedia); ok {
		return paid
	}
	return nil
}

// Video returns the document if message contains a video, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.Video(); doc != nil {
//		fmt.Printf("Video ID: %d\n", doc.ID)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) Video() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
			return doc
		}
	}
	return nil
}

// Audio returns the document if message contains an audio file, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.Audio(); doc != nil {
//		fmt.Printf("Audio ID: %d\n", doc.ID)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) Audio() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeAudio); ok {
			return doc
		}
	}
	return nil
}

// Voice returns the document if message contains a voice note, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.Voice(); doc != nil {
//		fmt.Printf("Voice note ID: %d\n", doc.ID)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) Voice() *Document {
	if m.Media == nil {
		return nil
	}
	media, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok || !media.Voice {
		return nil
	}
	if doc, ok := media.Document.(*tg.Document); ok {
		return NewDocument(doc)
	}
	return nil
}

// Animation returns the document if message contains an animation (GIF), nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.Animation(); doc != nil {
//		fmt.Printf("Animation ID: %d\n", doc.ID)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) Animation() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeAnimated); ok {
			return doc
		}
	}
	return nil
}

// VideoNote returns the document if message contains a video note (round video), nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.VideoNote(); doc != nil {
//		fmt.Printf("Video note ID: %d\n", doc.ID)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) VideoNote() *Document {
	if m.Media == nil {
		return nil
	}
	media, ok := m.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil
	}
	if !media.Round || !media.Video {
		return nil
	}
	if doc, ok := media.Document.(*tg.Document); ok {
		return NewDocument(doc)
	}
	return nil
}

// Sticker returns the document if message contains a sticker, nil otherwise.
// Returns a wrapper type with helper methods like FileID().
//
// Example:
//
//	if doc := m.Sticker(); doc != nil {
//		fmt.Printf("Sticker ID: %d\n", doc.ID)
//		fmt.Printf("File ID: %s\n", doc.FileID())
//	}
func (m *Message) Sticker() *Document {
	doc := m.Document()
	if doc == nil {
		return nil
	}
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeSticker); ok {
			return doc
		}
	}
	return nil
}

// IsMedia returns true if the message contains any media.
//
// Example:
//
//	if m.IsMedia() {
//		fmt.Printf("Message has media\n")
//	}
func (m *Message) IsMedia() bool {
	return m.Media != nil
}

// FileID returns a formatted file ID string for any media in the message.
// This is a convenience method that calls FileID() on the appropriate media wrapper.
// Returns empty string if the message contains no media.
//
// Example:
//
//	fileID := m.FileID()
//	if fileID != "" {
//		fmt.Printf("File ID: %s\n", fileID)
//	}
func (m *Message) FileID() string {
	// Check for photo media
	if photo := m.Photo(); photo != nil {
		return photo.FileID()
	}

	// Check for document media (including videos, audio, stickers, etc.)
	if doc := m.Document(); doc != nil {
		return doc.FileID()
	}

	return ""
}

// Link returns a clickable Telegram link to the message.
// Returns an empty string for private messages (no valid link format).
// For public channels/groups with username: https://t.me/username/msgID
// For private channels/groups: https://t.me/c/channelID/msgID
//
// Example:
//
//	link := m.Link()
//	if link != "" {
//		fmt.Printf("Message link: %s\n", link)
//	}
func (m *Message) Link() string {
	// No link possible without PeerID
	if m.PeerID == nil {
		return ""
	}

	chatID := functions.GetChatIDFromPeer(m.PeerID)
	if chatID == 0 {
		return ""
	}

	// Private chats have no public links
	if _, isUser := m.PeerID.(*tg.PeerUser); isUser {
		return ""
	}

	// Try public username link first
	peer := m.PeerStorage.GetPeerByID(chatID)
	if peer != nil && peer.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", peer.Username, m.ID)
	}

	// Private channel/group link using /c/ format
	// Strip the -100 prefix from channel IDs for URLs
	strippedID := stripChannelPrefix(chatID)
	return fmt.Sprintf("https://t.me/c/%d/%d", strippedID, m.ID)
}

// stripChannelPrefix removes the -100 prefix from megagroup/channel IDs
// for use in Telegram URLs.
func stripChannelPrefix(id int64) int64 {
	if id < 0 {
		id = -id
		// Channel IDs have -100 prefix in some contexts
		const channelPrefix = 1000000000000
		if id > channelPrefix {
			id -= channelPrefix
		}
	}
	return id
}
