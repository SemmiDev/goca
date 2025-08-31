package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	CharsetAlphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CharsetNumeric      = "0123456789"
	CharsetHex          = "abcdef0123456789"
)

// GenerateString returns a cryptographically secure random string
// using the provided charset.
func GenerateString(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than zero")
	}
	if len(charset) == 0 {
		return "", fmt.Errorf("charset must not be empty")
	}

	result := make([]byte, length)
	maxCharset := big.NewInt(int64(len(charset)))

	for i := range result {
		num, err := rand.Int(rand.Reader, maxCharset)
		if err != nil {
			return "", fmt.Errorf("failed to generate secure random number: %w", err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func String(length int) (string, error) {
	return GenerateString(length, CharsetAlphaNumeric)
}

func HexToken(length int) (string, error) {
	return GenerateString(length, CharsetHex)
}
