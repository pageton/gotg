package types

import (
	"fmt"

	"github.com/pageton/gotg/functions"
)

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
	return functions.UnPinMessage(m.Ctx, m.RawClient, m.PeerStorage, chatID, m.ID)
}

// UnpinAllMessages unpins all messages in this message's chat.
func (m *Message) UnpinAllMessages() error {
	if m.RawClient == nil {
		return fmt.Errorf("message has no client context")
	}
	chatID := functions.GetChatIDFromPeer(m.PeerID)
	return functions.UnPinAllMessages(m.Ctx, m.RawClient, m.PeerStorage, chatID)
}
