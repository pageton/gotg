package session

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/gotd/td/session"
)

// EncodeTelethonSession encodes session data into a Telethon-compatible session string.
//
// Telethon session strings use version prefix '1' followed by base64 URL-encoded
// binary data with big-endian byte order:
//
//	| Size |  Type  | Description |
//	|------|--------|-------------|
//	| 1    | byte   | DC ID       |
//	| 4/16 | bytes  | IP address  |
//	| 2    | uint16 | Port        |
//	| 256  | bytes  | Auth Key    |
//
// Parameters:
//   - data: The session data containing DC, Addr (ip:port), and AuthKey
//
// Returns the Telethon session string (prefixed with '1') or an error.
func EncodeTelethonSession(data *session.Data) (string, error) {
	if len(data.AuthKey) != 256 {
		return "", fmt.Errorf("telethon encode: auth key must be 256 bytes, got %d", len(data.AuthKey))
	}

	host, portStr, err := net.SplitHostPort(data.Addr)
	if err != nil {
		return "", fmt.Errorf("telethon encode: invalid address %q: %w", data.Addr, err)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("telethon encode: invalid IP address %q", host)
	}

	var port uint16
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return "", fmt.Errorf("telethon encode: invalid port %q: %w", portStr, err)
	}

	// Determine IPv4 or IPv6
	var ipBytes []byte
	if ip4 := ip.To4(); ip4 != nil {
		ipBytes = ip4
	} else {
		ipBytes = ip.To16()
	}

	// Total: 1 (DC) + len(ipBytes) + 2 (port) + 256 (key) = 263 or 275
	buf := make([]byte, 1+len(ipBytes)+2+256)

	// DC ID (1 byte)
	buf[0] = byte(data.DC) //nolint:gosec // G115: DC ID is 1-5

	// IP address (4 or 16 bytes)
	copy(buf[1:1+len(ipBytes)], ipBytes)

	// Port (2 bytes, big-endian)
	binary.BigEndian.PutUint16(buf[1+len(ipBytes):3+len(ipBytes)], port)

	// Auth Key (256 bytes)
	copy(buf[3+len(ipBytes):], data.AuthKey)

	// Version prefix '1' + base64 URL encoding
	return "1" + base64.URLEncoding.EncodeToString(buf), nil
}
