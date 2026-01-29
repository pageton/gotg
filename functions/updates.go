package functions

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// GetNewMessageUpdate extracts the new message from updates.
//
// Parameters:
//   - msgData: The message data to populate
//   - upds: The updates class to extract from
//   - p: Peer storage for resolving peer references
//
// Returns the new message or nil.
func GetNewMessageUpdate(msgData *tg.Message, upds tg.UpdatesClass, p *storage.PeerStorage) *tg.Message {
	u, ok := upds.(*tg.UpdateShortSentMessage)
	if ok {
		msgData.Flags = u.Flags
		msgData.Out = u.Out
		msgData.ID = u.ID
		msgData.Date = u.Date
		msgData.Media = u.Media
		msgData.Entities = u.Entities
		msgData.TTLPeriod = u.TTLPeriod
		return msgData
	}
	for _, update := range GetUpdateClassFromUpdatesClass(upds, p) {
		switch u := update.(type) {
		case *tg.UpdateNewMessage:
			return GetMessageFromMessageClass(u.Message)
		case *tg.UpdateNewChannelMessage:
			return GetMessageFromMessageClass(u.Message)
		case *tg.UpdateNewScheduledMessage:
			return GetMessageFromMessageClass(u.Message)
		case *tg.UpdateBotNewBusinessMessage:
			return GetMessageFromMessageClass(u.Message)
		}
	}
	return nil
}

// GetEditMessageUpdate extracts the edited message from updates.
//
// Parameters:
//   - upds: The updates class to extract from
//   - p: Peer storage for resolving peer references
//
// Returns the edited message or nil.
func GetEditMessageUpdate(upds tg.UpdatesClass, p *storage.PeerStorage) *tg.Message {
	for _, update := range GetUpdateClassFromUpdatesClass(upds, p) {
		switch u := update.(type) {
		case *tg.UpdateEditMessage:
			return GetMessageFromMessageClass(u.Message)
		case *tg.UpdateEditChannelMessage:
			return GetMessageFromMessageClass(u.Message)
		case *tg.UpdateBotEditBusinessMessage:
			return GetMessageFromMessageClass(u.Message)
		}
	}
	return nil
}

// GetUpdateClassFromUpdatesClass extracts update classes from updates.
//
// Parameters:
//   - updates: The updates class to extract from
//   - p: Peer storage for resolving peer references
//
// Returns list of update classes.
func GetUpdateClassFromUpdatesClass(updates tg.UpdatesClass, p *storage.PeerStorage) (u []tg.UpdateClass) {
	u, _, _ = getUpdateFromUpdates(updates, p)
	return
}

func getUpdateFromUpdates(updates tg.UpdatesClass, p *storage.PeerStorage) ([]tg.UpdateClass, []tg.ChatClass, []tg.UserClass) {
	switch u := updates.(type) {
	case *tg.Updates:
		SavePeersFromClassArray(p, u.Chats, u.Users)
		return u.Updates, u.Chats, u.Users
	case *tg.UpdatesCombined:
		SavePeersFromClassArray(p, u.Chats, u.Users)
		return u.Updates, u.Chats, u.Users
	case *tg.UpdateShort:
		return []tg.UpdateClass{u.Update}, tg.ChatClassArray{}, tg.UserClassArray{}
	default:
		return nil, nil, nil
	}
}

// GetMessageFromMessageClass extracts message from message class.
//
// Parameters:
//   - m: The message class to extract from
//
// Returns the message or nil.
func GetMessageFromMessageClass(m tg.MessageClass) *tg.Message {
	msg, ok := m.(*tg.Message)
	if !ok {
		return nil
	}
	return msg
}

// ReturnNewMessageWithError returns new message with error handling.
//
// Internal helper function.
//
// Parameters:
//   - msgData: The message data to populate
//   - upds: The updates class to extract from
//   - p: Peer storage for resolving peer references
//   - err: The error to check
//
// Returns the new message or error.
func ReturnNewMessageWithError(msgData *tg.Message, upds tg.UpdatesClass, p *storage.PeerStorage, err error) (*tg.Message, error) {
	if err != nil {
		return nil, err
	}
	if msgData == nil {
		msgData = &tg.Message{}
	}
	return GetNewMessageUpdate(msgData, upds, p), nil
}

// ReturnEditMessageWithError returns edited message with error handling.
//
// Internal helper function.
//
// Parameters:
//   - p: Peer storage for resolving peer references
//   - upds: The updates class to extract from
//   - err: The error to check
//
// Returns the edited message or error.
func ReturnEditMessageWithError(p *storage.PeerStorage, upds tg.UpdatesClass, err error) (*tg.Message, error) {
	if err != nil {
		return nil, err
	}
	return GetEditMessageUpdate(upds, p), nil
}
