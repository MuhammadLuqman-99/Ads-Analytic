package crypto

import (
	"strings"
	"testing"
)

func TestTokenEncryptor_NewTokenEncryptor(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		expectErr bool
	}{
		{
			name:      "valid 32 byte key",
			key:       "12345678901234567890123456789012",
			expectErr: false,
		},
		{
			name:      "key too short",
			key:       "shortkey",
			expectErr: true,
		},
		{
			name:      "key too long",
			key:       "123456789012345678901234567890123456789012345678901234567890",
			expectErr: true,
		},
		{
			name:      "empty key",
			key:       "",
			expectErr: true,
		},
		{
			name:      "31 byte key",
			key:       "1234567890123456789012345678901",
			expectErr: true,
		},
		{
			name:      "33 byte key",
			key:       "123456789012345678901234567890123",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := NewTokenEncryptor(tt.key)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if encryptor != nil {
					t.Error("expected nil encryptor when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if encryptor == nil {
					t.Error("expected non-nil encryptor")
				}
			}
		})
	}
}

func TestTokenEncryptor_EncryptDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple token",
			plaintext: "access_token_12345",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "long token",
			plaintext: strings.Repeat("a", 1000),
		},
		{
			name:      "special characters",
			plaintext: "token!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode characters",
			plaintext: "token_マレーシア_🚀_中文",
		},
		{
			name:      "JWT-like token",
			plaintext: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			// Verify ciphertext is different from plaintext (except empty)
			if tt.plaintext != "" && ciphertext == tt.plaintext {
				t.Error("ciphertext should be different from plaintext")
			}

			// Decrypt
			decrypted, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Verify round-trip
			if decrypted != tt.plaintext {
				t.Errorf("decrypted text does not match original: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestTokenEncryptor_UniqueNonce(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	plaintext := "test_token_for_nonce_uniqueness"
	ciphertexts := make(map[string]bool)

	// Encrypt the same plaintext multiple times
	for i := 0; i < 100; i++ {
		ciphertext, err := encryptor.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encryption failed: %v", err)
		}

		// Each ciphertext should be unique due to random nonce
		if ciphertexts[ciphertext] {
			t.Errorf("duplicate ciphertext generated at iteration %d", i)
		}
		ciphertexts[ciphertext] = true
	}
}

func TestTokenEncryptor_DecryptInvalidData(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext string
		expectErr  bool
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
			expectErr:  true,
		},
		{
			name:       "too short",
			ciphertext: "YWJj", // "abc" in base64
			expectErr:  true,
		},
		{
			name:       "tampered ciphertext",
			ciphertext: "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkw",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if tt.expectErr && err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}

func TestTokenEncryptor_WrongKey(t *testing.T) {
	key1 := "12345678901234567890123456789012"
	key2 := "abcdefghijklmnopqrstuvwxyz123456"

	encryptor1, _ := NewTokenEncryptor(key1)
	encryptor2, _ := NewTokenEncryptor(key2)

	plaintext := "secret_token"

	// Encrypt with key1
	ciphertext, err := encryptor1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Attempt to decrypt with key2
	_, err = encryptor2.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got: %v", err)
	}
}

func TestEncryptToken_ConvenienceFunction(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "test_access_token"

	encrypted, err := EncryptToken(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptToken failed: %v", err)
	}

	decrypted, err := DecryptToken(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptToken failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptToken_InvalidKey(t *testing.T) {
	invalidKey := "short"
	plaintext := "test_token"

	_, err := EncryptToken(plaintext, invalidKey)
	if err == nil {
		t.Error("expected error with invalid key")
	}
	if err != ErrInvalidKeyLength {
		t.Errorf("expected ErrInvalidKeyLength, got: %v", err)
	}
}

func BenchmarkTokenEncryptor_Encrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	encryptor, _ := NewTokenEncryptor(key)
	plaintext := "EAABwzLixnjYBO0ZCZByKZAR8ZCA0M8WZBx7ZA..."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryptor.Encrypt(plaintext)
	}
}

func BenchmarkTokenEncryptor_Decrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	encryptor, _ := NewTokenEncryptor(key)
	plaintext := "EAABwzLixnjYBO0ZCZByKZAR8ZCA0M8WZBx7ZA..."
	ciphertext, _ := encryptor.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryptor.Decrypt(ciphertext)
	}
}
