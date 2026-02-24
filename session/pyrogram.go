package session

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gotd/td/session"
)

type Key [256]byte

// ID returns auth_key_id.
func (k Key) ID() [8]byte {
	raw := sha1.Sum(k[:]) // #nosec
	var id [8]byte
	copy(id[:], raw[12:])
	return id
}

// WithID creates new AuthKey from Key.
func (k Key) WithID() AuthKey {
	return AuthKey{
		Value: k,
		ID:    k.ID(),
	}
}

// AuthKey is a Key with cached id.
type AuthKey struct {
	Value Key
	ID    [8]byte
}

func DecodePyrogramSession(hx string) (*session.Data, error) {
	if len(hx) < 1 {
		return nil, fmt.Errorf("given string too small: %d", len(hx))
	}

	hx = strings.TrimRight(hx, "=")
	if pad := len(hx) % 4; pad != 0 {
		hx += strings.Repeat("=", 4-pad)
	}
	data, err := base64.URLEncoding.DecodeString(hx)
	if err != nil {
		return nil, fmt.Errorf("decode hex: %w", err)
	}

	return decodeStringSession(data)
}

func decodeStringSession(data []byte) (*session.Data, error) {
	// Given parameter should contain version + data
	// where data encoded using pack as '>BI?256sQ?'
	// depending on IP type.
	//
	// Table:
	//
	// | Size |  Type  | Description |
	// |------|--------|-------------|
	// | 1    | int    | DC ID       |
	// | 4    | int    | APP ID      |
	// | 1    | bool   | Test Mode   |
	// | 256  | bytes  | Auth Key    |
	// | 8    | bytes  | User ID     |
	// | 1    | bool   | Is Bot      |
	if len(data) < 262 {
		return nil, fmt.Errorf("session data too short: expected at least 262 bytes, got %d", len(data))
	}

	dc := data[0]
	testMode := data[5] == 1
	var key Key
	copy(key[:], data[6:262])
	id := key.WithID().ID

	return &session.Data{
		DC:        int(dc),
		AuthKey:   key[:],
		AuthKeyID: id[:],
		Config: session.Config{
			TestMode: testMode,
		},
	}, nil
}
