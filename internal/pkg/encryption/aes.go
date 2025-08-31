package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// AES256CBC handles AES-256-CBC encryption/decryption (Laravel compatible)
type AES256CBC struct {
	key []byte
}

// EncryptedPayload is the JSON payload format used by Laravel
type EncryptedPayload struct {
	IV    string `json:"iv"`
	Value string `json:"value"`
	MAC   string `json:"mac"`
	Tag   string `json:"tag"`
}

// NewAES256CBC creates a new AES-256-CBC encryptor with the given key
// The key should be base64 encoded (like Laravel's APP_KEY)
func NewAES256CBC(appKey string) (*AES256CBC, error) {
	if strings.HasPrefix(strings.TrimSpace(appKey), "base64:") {
		appKey = strings.TrimPrefix(appKey, "base64:")
	}

	if appKey == "" {
		return nil, errors.New("app key cannot be empty")
	}

	key, err := base64.StdEncoding.DecodeString(appKey)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 key: %v", err)
	}

	if len(key) != 32 { // AES-256 requires 32 bytes
		return nil, errors.New("key must be 32 bytes (AES-256)")
	}

	return &AES256CBC{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-CBC compatible with Laravel
func (a *AES256CBC) Encrypt(plaintext string) (string, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	paddedText := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	payload := EncryptedPayload{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(ciphertext),
		MAC:   "",
		Tag:   "",
	}

	mac, err := a.generateMAC(payload)
	if err != nil {
		return "", fmt.Errorf("failed to generate MAC: %v", err)
	}
	payload.MAC = mac

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	return base64.StdEncoding.EncodeToString(jsonPayload), nil
}

// Decrypt decrypts Laravel encrypted data
func (a *AES256CBC) Decrypt(encryptedData string) (string, error) {
	jsonPayload, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("invalid base64 data: %v", err)
	}

	var payload EncryptedPayload
	if err := json.Unmarshal(jsonPayload, &payload); err != nil {
		return "", fmt.Errorf("invalid JSON payload: %v", err)
	}

	expectedMAC, err := a.generateMAC(payload)
	if err != nil {
		return "", fmt.Errorf("failed to generate MAC for verification: %v", err)
	}

	if !hmac.Equal([]byte(payload.MAC), []byte(expectedMAC)) {
		return "", errors.New("MAC verification failed - data may be tampered")
	}

	iv, err := base64.StdEncoding.DecodeString(payload.IV)
	if err != nil {
		return "", fmt.Errorf("invalid IV: %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Value)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %v", err)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	unpaddedText, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to remove padding: %v", err)
	}

	return string(unpaddedText), nil
}

// generateMAC creates HMAC-SHA256 compatible with Laravel
func (a *AES256CBC) generateMAC(payload EncryptedPayload) (string, error) {
	tempPayload := payload
	tempPayload.MAC = ""

	jsonData, err := json.Marshal(tempPayload)
	if err != nil {
		return "", err
	}

	macPayload := base64.StdEncoding.EncodeToString(jsonData)

	h := hmac.New(sha256.New, a.key)
	h.Write([]byte(macPayload))

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// pkcs7Pad pads data using PKCS7 padding
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS7 padding
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	if len(data)%blockSize != 0 {
		return nil, errors.New("data is not padded properly")
	}

	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, errors.New("invalid padding")
	}

	// Verify padding bytes
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padLen], nil
}
