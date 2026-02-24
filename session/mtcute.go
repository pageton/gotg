package session

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/session"
	gotgErrors "github.com/pageton/gotg/errors"
)

// DecodeMtcuteSession decodes an mtcute v3 string session into session.Data.
//
// mtcute session strings are URL-safe base64 encoded binary data using
// TL-style serialization with the following layout:
//
//	| Field            | Size     | Description                        |
//	|------------------|----------|------------------------------------|
//	| version          | 1 byte   | Must be 3                          |
//	| flags            | 4 bytes  | LE int32 (bit0=hasSelf, bit2=hasMedia) |
//	| primaryDC        | variable | TL bytes-wrapped DC option         |
//	| mediaDC          | variable | TL bytes-wrapped DC option (if hasMedia) |
//	| selfUserID       | 8 bytes  | LE int64 (if hasSelf)              |
//	| selfIsBot        | 4 bytes  | TL bool (if hasSelf)               |
//	| authKey          | variable | TL bytes-wrapped (256 bytes)       |
func DecodeMtcuteSession(sessionStr string) (*session.Data, error) {
	if len(sessionStr) == 0 {
		return nil, gotgErrors.ErrInvalidMtcuteSession
	}

	// mtcute uses URL-safe base64 with no padding
	sessionStr = strings.TrimRight(sessionStr, "=")
	if pad := len(sessionStr) % 4; pad != 0 {
		sessionStr += strings.Repeat("=", 4-pad)
	}
	buf, err := base64.URLEncoding.DecodeString(sessionStr)
	if err != nil {
		return nil, fmt.Errorf("mtcute: base64 decode: %w", err)
	}

	if len(buf) < 6 {
		return nil, gotgErrors.ErrInvalidMtcuteSession
	}

	version := buf[0]
	if version != 3 {
		return nil, fmt.Errorf("mtcute: unsupported version %d, expected 3", version)
	}

	r := &tlReader{data: buf, pos: 1}

	flags, err := r.readInt32()
	if err != nil {
		return nil, fmt.Errorf("mtcute: read flags: %w", err)
	}
	hasSelf := flags&1 != 0
	hasMedia := flags&4 != 0

	// Read primary DC option (TL bytes-wrapped)
	dcBytes, err := r.readTLBytes()
	if err != nil {
		return nil, fmt.Errorf("mtcute: read primary DC: %w", err)
	}
	primaryDC, err := parseMtcuteDcOption(dcBytes)
	if err != nil {
		return nil, fmt.Errorf("mtcute: parse primary DC: %w", err)
	}

	// Read media DC if present (we don't use it for session.Data but must skip it)
	if hasMedia {
		_, err := r.readTLBytes()
		if err != nil {
			return nil, fmt.Errorf("mtcute: read media DC: %w", err)
		}
	}

	// Read self info if present (skip for session.Data)
	if hasSelf {
		// int53: 8 bytes LE
		if err := r.skip(8); err != nil {
			return nil, fmt.Errorf("mtcute: skip user ID: %w", err)
		}
		// TL boolean: 4 bytes
		if err := r.skip(4); err != nil {
			return nil, fmt.Errorf("mtcute: skip isBot: %w", err)
		}
	}

	// Read auth key (TL bytes-wrapped)
	authKeyBytes, err := r.readTLBytes()
	if err != nil {
		return nil, fmt.Errorf("mtcute: read auth key: %w", err)
	}

	if len(authKeyBytes) != 256 {
		return nil, fmt.Errorf("mtcute: invalid auth key length: expected 256, got %d", len(authKeyBytes))
	}

	var key crypto.Key
	copy(key[:], authKeyBytes)
	id := key.WithID().ID

	return &session.Data{
		DC:        primaryDC.id,
		Addr:      net.JoinHostPort(primaryDC.ipAddress, strconv.Itoa(primaryDC.port)),
		AuthKey:   key[:],
		AuthKeyID: id[:],
	}, nil
}

// mtcuteDcOption represents a parsed DC option from mtcute format.
type mtcuteDcOption struct {
	id        int
	ipAddress string
	port      int
	ipv6      bool
	mediaOnly bool
	testMode  bool
}

// parseMtcuteDcOption parses a serialized BasicDcOption from mtcute.
//
// Format:
//
//	[version=2][dcId][flags] + TL string(ipAddress) + TL int32(port)
func parseMtcuteDcOption(data []byte) (*mtcuteDcOption, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("DC option too short: %d bytes", len(data))
	}

	version := data[0]
	if version != 1 && version != 2 {
		return nil, fmt.Errorf("unsupported DC option version: %d", version)
	}

	dcID := data[1]
	flags := data[2]

	r := &tlReader{data: data, pos: 3}

	// Read IP address as TL string (same encoding as TL bytes)
	ipBytes, err := r.readTLBytes()
	if err != nil {
		return nil, fmt.Errorf("read IP address: %w", err)
	}

	// Read port as int32 LE
	port, err := r.readInt32()
	if err != nil {
		return nil, fmt.Errorf("read port: %w", err)
	}

	opt := &mtcuteDcOption{
		id:        int(dcID),
		ipAddress: string(ipBytes),
		port:      int(port),
		ipv6:      flags&1 != 0,
		mediaOnly: flags&2 != 0,
	}
	if version == 2 {
		opt.testMode = flags&4 != 0
	}

	return opt, nil
}

// tlReader is a minimal TL binary reader (little-endian).
type tlReader struct {
	data []byte
	pos  int
}

func (r *tlReader) readInt32() (int32, error) {
	if r.pos+4 > len(r.data) {
		return 0, fmt.Errorf("read int32: need 4 bytes at pos %d, have %d", r.pos, len(r.data))
	}
	v := int32(binary.LittleEndian.Uint32(r.data[r.pos : r.pos+4])) //nolint:gosec // G115: intentional TL decoding
	r.pos += 4
	return v, nil
}

// readTLBytes reads TL-encoded bytes: length prefix + data + padding.
//
// If first byte < 254: length = firstByte, padding = (length+1) % 4
// If first byte == 254: length = next 3 bytes LE, padding = length % 4
func (r *tlReader) readTLBytes() ([]byte, error) {
	if r.pos >= len(r.data) {
		return nil, fmt.Errorf("readTLBytes: no data at pos %d", r.pos)
	}

	firstByte := r.data[r.pos]
	r.pos++

	var length int
	var padding int

	if firstByte == 254 {
		if r.pos+3 > len(r.data) {
			return nil, fmt.Errorf("readTLBytes: need 3 length bytes at pos %d", r.pos)
		}
		length = int(r.data[r.pos]) | int(r.data[r.pos+1])<<8 | int(r.data[r.pos+2])<<16
		r.pos += 3
		padding = length % 4
	} else {
		length = int(firstByte)
		padding = (length + 1) % 4
	}

	if r.pos+length > len(r.data) {
		return nil, fmt.Errorf("readTLBytes: need %d bytes at pos %d, have %d", length, r.pos, len(r.data))
	}

	result := make([]byte, length)
	copy(result, r.data[r.pos:r.pos+length])
	r.pos += length

	if padding > 0 {
		r.pos += 4 - padding
	}

	return result, nil
}

func (r *tlReader) skip(n int) error {
	if r.pos+n > len(r.data) {
		return fmt.Errorf("skip: need %d bytes at pos %d, have %d", n, r.pos, len(r.data))
	}
	r.pos += n
	return nil
}
