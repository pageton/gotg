package session

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gotd/td/session"
)

// EncodePyrogramSession encodes session data into a Pyrogram-compatible session string.
//
// Pyrogram session strings use the format '>BI?256sQ?' (big-endian packed):
//
//	| Size |  Type  | Description |
//	|------|--------|-------------|
//	| 1    | uint8  | DC ID       |
//	| 4    | uint32 | App ID      |
//	| 1    | bool   | Test Mode   |
//	| 256  | bytes  | Auth Key    |
//	| 8    | uint64 | User ID     |
//	| 1    | bool   | Is Bot      |
//
// Parameters:
//   - data: The session data containing DC, AuthKey, and Config
//   - appID: The Telegram API application ID
//   - userID: The authenticated user's Telegram ID
//   - isBot: Whether the session belongs to a bot
//
// Returns the base64 URL-encoded session string or an error.
func EncodePyrogramSession(data *session.Data, appID int32, userID int64, isBot bool) (string, error) {
	if len(data.AuthKey) != 256 {
		return "", fmt.Errorf("pyrogram encode: auth key must be 256 bytes, got %d", len(data.AuthKey))
	}

	buf := make([]byte, 271)

	// DC ID (1 byte)
	buf[0] = byte(data.DC) //nolint:gosec // G115: DC ID is 1-5

	// App ID (4 bytes, big-endian)
	binary.BigEndian.PutUint32(buf[1:5], uint32(appID)) //nolint:gosec // G115: app ID fits uint32

	// Test Mode (1 byte bool)
	if data.Config.TestMode {
		buf[5] = 1
	}

	// Auth Key (256 bytes)
	copy(buf[6:262], data.AuthKey)

	// User ID (8 bytes, big-endian)
	binary.BigEndian.PutUint64(buf[262:270], uint64(userID)) //nolint:gosec // G115: intentional int64->uint64

	// Is Bot (1 byte bool)
	if isBot {
		buf[270] = 1
	}

	encoded := base64.URLEncoding.EncodeToString(buf)
	encoded = strings.TrimRight(encoded, "=")
	return encoded, nil
}
