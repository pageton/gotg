package types

import (
	"fmt"
	"reflect"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/parsemode"
)

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
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
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
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
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

	var caption string
	var ents []tg.MessageEntityClass
	var markup tg.ReplyMarkupClass
	var parseMode string

	if len(opts) > 0 && opts[0] != nil {
		if e, ok := opts[0].([]tg.MessageEntityClass); ok {
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
