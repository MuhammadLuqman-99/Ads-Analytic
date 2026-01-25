package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrInvalidKeyLength is returned when the encryption key is not 32 bytes
	ErrInvalidKeyLength = errors.New("encryption key must be 32 bytes for AES-256")
	// ErrCiphertextTooShort is returned when the ciphertext is too short to be valid
	ErrCiphertextTooShort = errors.New("ciphertext too short")
	// ErrDecryptionFailed is returned when decryption fails (e.g., wrong key or tampered data)
	ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or key")
)

// TokenEncryptor handles encryption and decryption of OAuth tokens using AES-256-GCM
type TokenEncryptor struct {
	key []byte
}

// NewTokenEncryptor creates a new TokenEncryptor with the given key
// The key must be exactly 32 bytes for AES-256
func NewTokenEncryptor(key string) (*TokenEncryptor, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, ErrInvalidKeyLength
	}
	return &TokenEncryptor{key: keyBytes}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext
// The nonce is prepended to the ciphertext before encoding
func (e *TokenEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and prepend nonce to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64-encoded result
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext encrypted with AES-256-GCM
// Expects the nonce to be prepended to the ciphertext
func (e *TokenEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	// Extract nonce and ciphertext
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// EncryptToken is a convenience function for encrypting a single token
func EncryptToken(token, key string) (string, error) {
	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		return "", err
	}
	return encryptor.Encrypt(token)
}

// DecryptToken is a convenience function for decrypting a single token
func DecryptToken(encryptedToken, key string) (string, error) {
	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		return "", err
	}
	return encryptor.Decrypt(encryptedToken)
}
