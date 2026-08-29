package share

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const tokenEncryptionKeySize = 32

// ParseTokenEncryptionKey parses the base64-encoded AES-256 key used to
// protect the recoverable share token stored for owner-facing listings.
func ParseTokenEncryptionKey(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("share token encryption key is required")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode share token encryption key: %w", err)
	}
	if len(key) != tokenEncryptionKeySize {
		return nil, fmt.Errorf("share token encryption key must decode to %d bytes", tokenEncryptionKeySize)
	}
	return key, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}

// TokenHash returns the deterministic lookup hash for a share token.
func TokenHash(token string) string {
	return hashToken(token)
}

// NewToken generates an opaque bearer token for a new share.
func NewToken() (string, error) {
	return newToken()
}

// EncryptToken protects a token for durable owner-facing storage.
func EncryptToken(key []byte, token string) ([]byte, error) {
	return encryptToken(key, token)
}

// DecryptToken recovers a token from durable owner-facing storage.
func DecryptToken(key, encrypted []byte) (string, error) {
	return decryptToken(key, encrypted)
}

func encryptToken(key []byte, token string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create token cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create token encryption: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate token encryption nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, []byte(token), nil), nil
}

func decryptToken(key, encrypted []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create token cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create token encryption: %w", err)
	}
	if len(encrypted) < gcm.NonceSize() {
		return "", errors.New("encrypted share token is truncated")
	}
	nonce, ciphertext := encrypted[:gcm.NonceSize()], encrypted[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt share token: %w", err)
	}
	return string(plaintext), nil
}
