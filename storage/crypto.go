package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// SessionEncryptor provides AES-256-GCM encryption for session data at rest.
// It protects the 256-byte MTProto auth key stored in the session database
// from being readable by anyone with direct database access.
//
// Create with NewSessionEncryptor. The key must be exactly 32 bytes and
// should be sourced from a secure secret store (environment variable,
// /run/secrets, KMS, etc.).
type SessionEncryptor struct {
	aead cipher.AEAD
}

// NewSessionEncryptor creates a SessionEncryptor from a 32-byte key.
func NewSessionEncryptor(key []byte) (*SessionEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New("storage: session encryption key must be exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SessionEncryptor{aead: aead}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// The output format is: nonce (12 bytes) || ciphertext || tag (16 bytes).
func (e *SessionEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return e.aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data produced by Encrypt.
func (e *SessionEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := e.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("storage: session ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return e.aead.Open(nil, nonce, ct, nil)
}
