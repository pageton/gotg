package session

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/gotd/td/session"
)

// EncodeMtcuteSession encodes session data into an mtcute v3-compatible session string.
//
// mtcute session strings are URL-safe base64 encoded binary data using
// TL-style serialization with the following layout:
//
//	| Field            | Size     | Description                                    |
//	|------------------|----------|------------------------------------------------|
//	| version          | 1 byte   | Always 3                                       |
//	| flags            | 4 bytes  | LE int32 (bit0=hasSelf, bit2=hasMedia)         |
//	| primaryDC        | variable | TL bytes-wrapped DC option                     |
//	| selfUserID       | 8 bytes  | LE int64 (if hasSelf)                          |
//	| selfIsBot        | 4 bytes  | TL bool (if hasSelf)                           |
//	| authKey          | variable | TL bytes-wrapped (256 bytes)                   |
//
// Parameters:
//   - data: The session data containing DC, Addr (ip:port), and AuthKey
//   - userID: The authenticated user's Telegram ID (0 to omit self info)
//   - isBot: Whether the session belongs to a bot
//
// Returns the URL-safe base64-encoded session string or an error.
func EncodeMtcuteSession(data *session.Data, userID int64, isBot bool) (string, error) {
	if len(data.AuthKey) != 256 {
		return "", fmt.Errorf("mtcute encode: auth key must be 256 bytes, got %d", len(data.AuthKey))
	}

	host, portStr, err := net.SplitHostPort(data.Addr)
	if err != nil {
		return "", fmt.Errorf("mtcute encode: invalid address %q: %w", data.Addr, err)
	}

	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return "", fmt.Errorf("mtcute encode: invalid port %q: %w", portStr, err)
	}

	ip := net.ParseIP(host)
	isIPv6 := ip != nil && ip.To4() == nil

	w := &tlWriter{}

	// Version byte
	w.writeByte(3)

	// Flags
	var flags int32
	hasSelf := userID != 0
	if hasSelf {
		flags |= 1
	}
	// We don't write a separate media DC (bit 2 not set)
	w.writeInt32(flags)

	// Primary DC option as TL bytes
	dcOption := serializeMtcuteDcOption(data.DC, host, port, isIPv6, false, data.Config.TestMode)
	w.writeTLBytes(dcOption)

	// Self info (if present)
	if hasSelf {
		w.writeInt53(userID)
		w.writeTLBool(isBot)
	}

	// Auth key as TL bytes
	w.writeTLBytes(data.AuthKey)

	encoded := base64.URLEncoding.EncodeToString(w.buf)
	encoded = strings.TrimRight(encoded, "=")
	return encoded, nil
}

// serializeMtcuteDcOption serializes a DC option in mtcute's BasicDcOption format.
//
// Format: [version=2][dcId][flags] + TL string(ipAddress) + TL int32(port)
func serializeMtcuteDcOption(dcID int, ipAddress string, port int, ipv6, mediaOnly, testMode bool) []byte {
	w := &tlWriter{}

	var flags byte
	if ipv6 {
		flags |= 1
	}
	if mediaOnly {
		flags |= 2
	}
	if testMode {
		flags |= 4
	}

	// Raw header: version, dcId, flags
	w.writeByte(2) // version
	w.writeByte(byte(dcID)) //nolint:gosec // G115: DC ID is 1-5
	w.writeByte(flags)

	// IP address as TL string
	w.writeTLBytes([]byte(ipAddress))

	// Port as int32 LE
	w.writeInt32(int32(port)) //nolint:gosec // G115: port fits int32

	return w.buf
}

// tlWriter is a minimal TL binary writer (little-endian).
type tlWriter struct {
	buf []byte
}

func (w *tlWriter) writeByte(b byte) {
	w.buf = append(w.buf, b)
}

func (w *tlWriter) writeInt32(v int32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v)) //nolint:gosec // G115: intentional TL encoding
	w.buf = append(w.buf, b...)
}

// writeInt53 writes a 64-bit integer in TL int53 format (two int32 LE: low then high).
func (w *tlWriter) writeInt53(v int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v)) //nolint:gosec // G115: intentional TL encoding
	w.buf = append(w.buf, b...)
}

// writeTLBool writes a TL-encoded boolean.
// true = 0x997275B5, false = 0xBC799737
func (w *tlWriter) writeTLBool(v bool) {
	b := make([]byte, 4)
	if v {
		binary.LittleEndian.PutUint32(b, 0x997275B5)
	} else {
		binary.LittleEndian.PutUint32(b, 0xBC799737)
	}
	w.buf = append(w.buf, b...)
}

// writeTLBytes writes TL-encoded bytes: length prefix + data + padding.
//
// If length <= 253: [1 byte length] [data] [padding to 4-byte boundary]
// If length > 253: [0xFE] [3 bytes LE length] [data] [padding to 4-byte boundary]
func (w *tlWriter) writeTLBytes(data []byte) {
	length := len(data)
	var padding int

	if length <= 253 {
		w.buf = append(w.buf, byte(length))
		padding = (length + 1) % 4
	} else {
		w.buf = append(w.buf, 254)
		w.buf = append(w.buf, byte(length&0xFF), byte((length>>8)&0xFF), byte((length>>16)&0xFF))
		padding = length % 4
	}

	w.buf = append(w.buf, data...)

	if padding > 0 {
		w.buf = append(w.buf, make([]byte, 4-padding)...)
	}
}
