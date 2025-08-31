package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		name        string
		appKey      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid base64 key with prefix",
			appKey:      "base64:" + base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expectError: false,
		},
		{
			name:        "Valid base64 key without prefix",
			appKey:      base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expectError: false,
		},
		{
			name:        "Invalid base64 key",
			appKey:      "invalid-base64-key!@#",
			expectError: true,
			errorMsg:    "invalid base64 key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := NewAES256CBC(tt.appKey)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if encryptor != nil {
					t.Errorf("Expected nil encryptor on error, got: %v", encryptor)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if encryptor == nil {
					t.Errorf("Expected valid encryptor, got nil")
				}
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// Generate a valid key for testing
	key := make([]byte, 32)
	rand.Read(key)
	appKey := "base64:" + base64.StdEncoding.EncodeToString(key)

	encryptor, err := NewAES256CBC(appKey)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "Simple text",
			plaintext: "Hello World",
		},
		{
			name:      "Empty string",
			plaintext: "",
		},
		{
			name:      "Long text",
			plaintext: strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 100),
		},
		{
			name:      "Unicode text",
			plaintext: "Hello 世界! 🌍 Привет мир! مرحبا بالعالم!",
		},
		{
			name:      "JSON data",
			plaintext: `{"name":"John","age":30,"city":"New York","active":true}`,
		},
		{
			name:      "Special characters",
			plaintext: "!@#$%^&*()_+{}|:<>?[]\\;'\",./-=`~",
		},
		{
			name:      "Newlines and tabs",
			plaintext: "Line 1\nLine 2\tTabbed\r\nWindows line ending",
		},
		{
			name:      "Very short",
			plaintext: "x",
		},
		{
			name:      "Exactly block size (16 chars)",
			plaintext: "1234567890123456",
		},
		{
			name:      "Block size + 1 (17 chars)",
			plaintext: "12345678901234567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Errorf("Encryption failed: %v", err)
				return
			}

			if encrypted == "" {
				t.Errorf("Encrypted string is empty")
				return
			}

			// Ensure encrypted data is different from plaintext
			if encrypted == tt.plaintext {
				t.Errorf("Encrypted data should be different from plaintext")
			}

			// Test decryption
			decrypted, err := encryptor.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decryption failed: %v", err)
				return
			}

			// Verify decrypted matches original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text doesn't match original.\nExpected: %q\nGot: %q", tt.plaintext, decrypted)
			}
		})
	}
}
