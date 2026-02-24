package session

import (
	"crypto/rand"
	"testing"

	"github.com/gotd/td/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestAuthKey() []byte {
	key := make([]byte, 256)
	_, _ = rand.Read(key)
	return key
}

func makeTestSessionData() *session.Data {
	key := makeTestAuthKey()
	var k Key
	copy(k[:], key)
	id := k.WithID().ID

	return &session.Data{
		DC:        2,
		Addr:      "149.154.167.50:443",
		AuthKey:   key,
		AuthKeyID: id[:],
	}
}

func TestPyrogramRoundTrip(t *testing.T) {
	original := makeTestSessionData()
	appID := int32(12345)
	userID := int64(987654321)
	isBot := false

	encoded, err := EncodePyrogramSession(original, appID, userID, isBot)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	decoded, err := DecodePyrogramSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
	assert.Equal(t, original.AuthKeyID, decoded.AuthKeyID)
}

func TestTelethonRoundTrip(t *testing.T) {
	original := makeTestSessionData()

	encoded, err := EncodeTelethonSession(original)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	// Telethon decoding is handled by gotd/td's session.TelethonSession
	decoded, err := session.TelethonSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.Addr, decoded.Addr)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
	assert.Equal(t, original.AuthKeyID, decoded.AuthKeyID)
}

func TestGramjsRoundTrip(t *testing.T) {
	original := makeTestSessionData()

	encoded, err := EncodeGramjsSession(original)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	decoded, err := DecodeGramjsSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.Addr, decoded.Addr)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
	assert.Equal(t, original.AuthKeyID, decoded.AuthKeyID)
}

func TestMtcuteRoundTrip(t *testing.T) {
	original := makeTestSessionData()
	userID := int64(987654321)
	isBot := false

	encoded, err := EncodeMtcuteSession(original, userID, isBot)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	decoded, err := DecodeMtcuteSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.Addr, decoded.Addr)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
	assert.Equal(t, original.AuthKeyID, decoded.AuthKeyID)
}

func TestMtcuteRoundTripWithBot(t *testing.T) {
	original := makeTestSessionData()
	userID := int64(123456789)
	isBot := true

	encoded, err := EncodeMtcuteSession(original, userID, isBot)
	require.NoError(t, err)

	decoded, err := DecodeMtcuteSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.Addr, decoded.Addr)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
}

func TestMtcuteRoundTripNoSelf(t *testing.T) {
	original := makeTestSessionData()

	// userID=0 means no self info
	encoded, err := EncodeMtcuteSession(original, 0, false)
	require.NoError(t, err)

	decoded, err := DecodeMtcuteSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.Addr, decoded.Addr)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
}

func TestMtcuteDecodeRealSession(t *testing.T) {
	// This is the test vector from mtcute's test suite (base.test.ts)
	// It's a stub client session with zeroed auth key
	sessionStr := "AwQAAAAXAgIADjE0OS4xNTQuMTY3LjUwALsBAAAXAgICDzE0OS4xNTQuMTY3LjIyMrsBAAD-AAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	decoded, err := DecodeMtcuteSession(sessionStr)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	assert.Equal(t, 2, decoded.DC)
	assert.Contains(t, decoded.Addr, "149.154.167.50")
}

func TestMtcuteDecodeTooShort(t *testing.T) {
	_, err := DecodeMtcuteSession("")
	assert.Error(t, err)
}

func TestMtcuteDecodeInvalidVersion(t *testing.T) {
	// Version 1 instead of 3
	_, err := DecodeMtcuteSession("AQ==")
	assert.Error(t, err)
}

func TestPyrogramEncodeInvalidKeyLength(t *testing.T) {
	data := &session.Data{AuthKey: make([]byte, 128)}
	_, err := EncodePyrogramSession(data, 0, 0, false)
	assert.Error(t, err)
}

func TestTelethonEncodeInvalidKeyLength(t *testing.T) {
	data := &session.Data{AuthKey: make([]byte, 128)}
	_, err := EncodeTelethonSession(data)
	assert.Error(t, err)
}

func TestGramjsEncodeInvalidKeyLength(t *testing.T) {
	data := &session.Data{AuthKey: make([]byte, 128)}
	_, err := EncodeGramjsSession(data)
	assert.Error(t, err)
}

func TestMtcuteEncodeInvalidKeyLength(t *testing.T) {
	data := &session.Data{AuthKey: make([]byte, 128)}
	_, err := EncodeMtcuteSession(data, 0, false)
	assert.Error(t, err)
}

func TestTelethonIPv6RoundTrip(t *testing.T) {
	key := makeTestAuthKey()
	var k Key
	copy(k[:], key)
	id := k.WithID().ID

	original := &session.Data{
		DC:        1,
		Addr:      "[2001:67c:4e8:f002::e]:443",
		AuthKey:   key,
		AuthKeyID: id[:],
	}

	encoded, err := EncodeTelethonSession(original)
	require.NoError(t, err)

	decoded, err := session.TelethonSession(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.DC, decoded.DC)
	assert.Equal(t, original.AuthKey, decoded.AuthKey)
}
