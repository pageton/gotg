package session

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/gotd/td/session"
)

// EncodeGramjsSession encodes session data into a GramJS-compatible session string.
//
// GramJS session strings use version prefix '1' followed by base64 standard-encoded
// binary data with big-endian byte order:
//
//	| Size     |  Type  | Description       |
//	|----------|--------|-------------------|
//	| 1        | uint8  | DC ID             |
//	| 2        | uint16 | Address length    |
//	| variable | bytes  | Server address    |
//	| 2        | uint16 | Port              |
//	| 256      | bytes  | Auth Key          |
//
// Parameters:
//   - data: The session data containing DC, Addr (ip:port), and AuthKey
//
// Returns the GramJS session string (prefixed with '1') or an error.
func EncodeGramjsSession(data *session.Data) (string, error) {
	if len(data.AuthKey) != 256 {
		return "", fmt.Errorf("gramjs encode: auth key must be 256 bytes, got %d", len(data.AuthKey))
	}

	host, portStr, err := net.SplitHostPort(data.Addr)
	if err != nil {
		return "", fmt.Errorf("gramjs encode: invalid address %q: %w", data.Addr, err)
	}

	var port uint16
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return "", fmt.Errorf("gramjs encode: invalid port %q: %w", portStr, err)
	}

	addrBytes := []byte(host)
	addrLen := len(addrBytes)

	// Total: 1 (DCID) + 2 (addrLen) + addrLen + 2 (port) + 256 (key)
	buf := make([]byte, 1+2+addrLen+2+256)

	// DC ID (1 byte)
	buf[0] = byte(data.DC) //nolint:gosec // G115: DC ID is 1-5

	// Address length (2 bytes, big-endian)
	binary.BigEndian.PutUint16(buf[1:3], uint16(addrLen)) //nolint:gosec // G115: address length fits uint16

	// Address (variable)
	copy(buf[3:3+addrLen], addrBytes)

	// Port (2 bytes, big-endian)
	binary.BigEndian.PutUint16(buf[3+addrLen:5+addrLen], port)

	// Auth Key (256 bytes)
	copy(buf[5+addrLen:], data.AuthKey)

	// Version prefix '1' + base64 standard encoding
	return "1" + base64.StdEncoding.EncodeToString(buf), nil
}
