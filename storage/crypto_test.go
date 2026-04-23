package storage

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestKey() []byte {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return key
}

func TestSessionEncryptorRoundTrip(t *testing.T) {
	key := makeTestKey()
	enc, err := NewSessionEncryptor(key)
	require.NoError(t, err)

	plaintext := make([]byte, 256) // simulates a 256-byte auth key
	_, _ = rand.Read(plaintext)

	ciphertext, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	// Ciphertext must be different from plaintext and longer (nonce + tag).
	assert.NotEqual(t, plaintext, ciphertext)
	assert.Greater(t, len(ciphertext), len(plaintext))

	decrypted, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestSessionEncryptorRoundTripEmptyData(t *testing.T) {
	key := makeTestKey()
	enc, err := NewSessionEncryptor(key)
	require.NoError(t, err)

	// Empty data is a valid input — the encryption layer only encrypts
	// non-empty data (see session.go), but the encryptor itself should
	// handle it correctly.
	ciphertext, err := enc.Encrypt([]byte{})
	require.NoError(t, err)
	// AES-GCM on empty plaintext produces nonce + tag (28 bytes).
	assert.Equal(t, 28, len(ciphertext))

	decrypted, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, 0, len(decrypted))
}

func TestSessionEncryptorWrongKey(t *testing.T) {
	key1 := makeTestKey()
	key2 := makeTestKey()

	enc1, err := NewSessionEncryptor(key1)
	require.NoError(t, err)
	enc2, err := NewSessionEncryptor(key2)
	require.NoError(t, err)

	plaintext := []byte("sensitive auth key data")
	ciphertext, err := enc1.Encrypt(plaintext)
	require.NoError(t, err)

	_, err = enc2.Decrypt(ciphertext)
	assert.Error(t, err, "decrypting with wrong key should fail")
}

func TestSessionEncryptorTamperedCiphertext(t *testing.T) {
	key := makeTestKey()
	enc, err := NewSessionEncryptor(key)
	require.NoError(t, err)

	ciphertext, err := enc.Encrypt([]byte("test data"))
	require.NoError(t, err)

	// Flip a bit in the ciphertext.
	ciphertext[len(ciphertext)-1] ^= 0x01

	_, err = enc.Decrypt(ciphertext)
	assert.Error(t, err, "decrypting tampered ciphertext should fail")
}

func TestSessionEncryptorTruncatedCiphertext(t *testing.T) {
	key := makeTestKey()
	enc, err := NewSessionEncryptor(key)
	require.NoError(t, err)

	_, err = enc.Decrypt([]byte{0x01, 0x02}) // too short
	assert.Error(t, err)
}

func TestSessionEncryptorInvalidKeySize(t *testing.T) {
	_, err := NewSessionEncryptor([]byte("short"))
	assert.Error(t, err)

	_, err = NewSessionEncryptor(make([]byte, 64))
	assert.Error(t, err)

	_, err = NewSessionEncryptor(make([]byte, 16))
	assert.Error(t, err)
}

func TestSessionEncryptorDeterministicPlaintextProducesDifferentCiphertexts(t *testing.T) {
	key := makeTestKey()
	enc, err := NewSessionEncryptor(key)
	require.NoError(t, err)

	plaintext := []byte("same data encrypted twice")

	ct1, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	ct2, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	// Random nonce means identical plaintext produces different ciphertexts.
	assert.NotEqual(t, ct1, ct2)

	// But both decrypt to the same plaintext.
	d1, err := enc.Decrypt(ct1)
	require.NoError(t, err)
	d2, err := enc.Decrypt(ct2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, d1)
	assert.Equal(t, plaintext, d2)
}
