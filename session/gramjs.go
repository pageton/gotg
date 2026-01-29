package session

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"net"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/session"
)

func DecodeGramjsSession(hx string) (*session.Data, error) {
	return decodeGramjsSession(hx)
}

func decodeGramjsSession(sessionStr string) (*session.Data, error) {
	data := struct {
		Version       string
		DCID          uint8
		ServerAddress string
		Port          int16
		Key           []byte
		AuthKey       string
		KeyID         string
	}{}

	if len(sessionStr) < 2 || sessionStr[0] != '1' {
		return nil, errors.New("invalid session string: too short or wrong version")
	}
	strsession := sessionStr[1:]
	decodedBytes, err := base64.StdEncoding.DecodeString(strsession)
	if err != nil {
		return nil, errors.Wrap(err, "invalid base64 encoding")
	}

	if len(decodedBytes) < 71 {
		return nil, errors.Errorf("session data too short: expected at least 71 bytes, got %d", len(decodedBytes))
	}

	reader := bytes.NewReader(decodedBytes)

	data.Version = "1"

	err = binary.Read(reader, binary.BigEndian, &data.DCID)
	if err != nil {
		return nil, err
	}

	addressLength := make([]byte, 2)
	_, err = reader.Read(addressLength)
	if err != nil {
		return nil, err
	}

	addressLen := int(binary.BigEndian.Uint16(addressLength))
	addressBuffer := make([]byte, addressLen)
	_, err = reader.Read(addressBuffer)
	if err != nil {
		return nil, err
	}
	data.ServerAddress = string(bytes.TrimRight(addressBuffer, "\x00"))

	portBuffer := make([]byte, 2)
	_, err = reader.Read(portBuffer)
	if err != nil {
		return nil, err
	}
	data.Port = int16(binary.BigEndian.Uint16(portBuffer))

	keyLength := len(decodedBytes) - (5 + addressLen) // DCID(1 byte), Address Length(2 bytes), Address, Port(2 bytes)
	data.Key = make([]byte, keyLength)
	_, err = reader.Read(data.Key)
	if err != nil {
		return nil, err
	}

	data.AuthKey = base64.StdEncoding.EncodeToString(data.Key)
	keyid := crypto.Key(data.Key).WithID().ID
	data.KeyID = base64.StdEncoding.EncodeToString(keyid[:])
	return &session.Data{
		DC:        int(data.DCID),
		Addr:      net.JoinHostPort(data.ServerAddress, strconv.Itoa(int(data.Port))),
		AuthKey:   data.Key,
		AuthKeyID: keyid[:],
	}, nil
}
