package types

import (
	"fmt"
	"reflect"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

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
//	fileID := photoMsg.FileID()
//	newMsg, err := originalMsg.ReplyMediaWithFileID(fileID, "Here's the media you requested")
func (m *Message) ReplyMediaWithFileID(fileID string, caption string, opts ...any) (*Message, error) {
	media, err := InputMediaFromFileID(fileID, caption)
	if err != nil {
		return nil, fmt.Errorf("invalid fileID: %w", err)
	}
	return m.ReplyMedia(media, caption, opts...)
}
