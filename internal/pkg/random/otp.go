package random

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/sammidev/goca/internal/pkg/apperror"
)

const (
	minOTPLength = 4
	maxOTPLength = 12
	digits       = "0123456789"
)

// GenerateNumericOTP menghasilkan OTP numerik (digit-only) dengan panjang tertentu.
// Panjang valid antara minOTPLength dan maxOTPLength.
func GenerateNumericOTP(length int) (string, error) {
	if length < minOTPLength || length > maxOTPLength {
		return "", apperror.NewAppError(
			apperror.ErrCodeInvalidInput,
			"OTP length must be between 4 and 12 digits",
		)
	}

	digitsLength := big.NewInt(int64(len(digits)))
	var otpBuilder strings.Builder
	otpBuilder.Grow(length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, digitsLength)
		if err != nil {
			return "", apperror.NewAppErrorWithCause(
				apperror.ErrCodeInternalError,
				"Failed to generate random OTP",
				err,
			)
		}
		otpBuilder.WriteByte(digits[n.Int64()])
	}

	return otpBuilder.String(), nil
}
