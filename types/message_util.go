package types

import (
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// GetUser fetches and returns the full user info of the message sender.
// Returns nil if the message has no sender or on error.
//
// Example:
//
//	user, err := u.EffectiveMessage.GetUser()
//	user, err := u.EffectiveReply().GetUser()
func (m *Message) GetUser() (*tg.User, error) {
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

// GetFullUser fetches and returns the full user info of the message sender.
// Returns nil if the message has no sender or on error.
//
// Example:
//
//	user, err := u.EffectiveMessage.GetFullUser()
//	user, err := u.EffectiveReply().GetFullUser()
func (m *Message) GetFullUser() (*tg.UserFull, error) {
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

	return functions.GetFullUser(m.Ctx, m.RawClient, m.PeerStorage, peerUser.UserID)
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
	if m.PeerID == nil {
		return ""
	}

	chatID := functions.GetChatIDFromPeer(m.PeerID)
	if chatID == 0 {
		return ""
	}

	if _, isUser := m.PeerID.(*tg.PeerUser); isUser {
		return ""
	}

	peer := m.PeerStorage.GetPeerByID(chatID)
	if peer != nil && peer.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", peer.Username, m.ID)
	}

	strippedID := stripChannelPrefix(chatID)
	return fmt.Sprintf("https://t.me/c/%d/%d", strippedID, m.ID)
}

func stripChannelPrefix(id int64) int64 {
	if id < 0 {
		id = -id
		const channelPrefix = 1000000000000
		if id > channelPrefix {
			id -= channelPrefix
		}
	}
	return id
}
