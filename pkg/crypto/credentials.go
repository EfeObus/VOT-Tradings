// Package crypto provides AES-256-GCM encryption for data at rest — used to
// store each user's brokerage credentials in Postgres without keeping them
// in plaintext.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Box encrypts and decrypts small secrets (broker API keys/tokens) with a
// single symmetric key held only by the gateway process.
type Box struct {
	gcm cipher.AEAD
}

// NewBox builds a Box from a base64-encoded 32-byte key (see
// GenerateKeyBase64 to create one for local dev).
func NewBox(keyBase64 string) (*Box, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("crypto: decode key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 32 bytes after base64 decoding, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	return &Box{gcm: gcm}, nil
}

// Encrypt returns nonce||ciphertext, ready to store as a single BYTEA.
func (b *Box) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, b.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: read nonce: %w", err)
	}
	return b.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt reverses Encrypt.
func (b *Box) Decrypt(data []byte) ([]byte, error) {
	nonceSize := b.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("crypto: ciphertext shorter than nonce")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := b.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decrypt: %w", err)
	}
	return plaintext, nil
}

// GenerateKeyBase64 returns a fresh random 32-byte key, base64-encoded —
// use this once to populate CREDENTIAL_ENCRYPTION_KEY in .env.
func GenerateKeyBase64() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("crypto: generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
