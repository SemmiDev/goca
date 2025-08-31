package apperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sammidev/goca/internal/pkg/validator"
)

// ErrorCode represents modules-specific error codes
type ErrorCode string

const (
	// Domain-specific error codes
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeDatabaseError    ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService  ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"

	// Business logic error codes
	ErrCodeUserAlreadyExists     ErrorCode = "USER_ALREADY_EXISTS"
	ErrCodeUserNotFound          ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserIncorrectPassword ErrorCode = "USER_INCORRECT_PASSWORD"
	ErrCodeUserInactive          ErrorCode = "USER_INACTIVE"
	ErrCodeUserEmailNotVerified  ErrorCode = "USER_EMAIL_NOT_VERIFIED"
	ErrCodeInvalidOTP            ErrorCode = "INVALID_OTP"
	ErrCodeOTPExpired            ErrorCode = "OTP_EXPIRED"
	ErrCodeInvalidToken          ErrorCode = "INVALID_TOKEN"
)

func (e ErrorCode) String() string {
	return string(e)
}

// AppError represents an modules-specific error
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
}

// Error implements the error interfaces
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *AppError) Is(target error) bool {
	var appErr *AppError
	if errors.As(target, &appErr) {
		return e.Code == appErr.Code
	}
	return false
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (e *AppError) HTTPStatusCode() int {
	switch e.Code {
	case ErrCodeNotFound, ErrCodeUserNotFound:
		return http.StatusNotFound
	case ErrCodeInvalidInput, ErrCodeValidationFailed, ErrCodeInvalidOTP:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeUserIncorrectPassword, ErrCodeUserInactive, ErrCodeUserEmailNotVerified, ErrCodeInvalidToken:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict, ErrCodeUserAlreadyExists:
		return http.StatusConflict
	case ErrCodeExternalService:
		return http.StatusServiceUnavailable
	case ErrCodeDatabaseError, ErrCodeInternalError, ErrCodeOTPExpired:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ValidationErrors represents a collection of validation apperror
type ValidationErrors struct {
	*AppError
	Fields validator.ValidationErrors `json:"fields"`
}

// NewAppError creates a new modules error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithCause creates a new modules error with underlying cause
func NewAppErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewAppErrorWithDetails creates a new modules error with details
func NewAppErrorWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func NewValidationError(fields validator.ValidationErrors) *ValidationErrors {
	return &ValidationErrors{
		AppError: &AppError{
			Code:    ErrCodeValidationFailed,
			Message: "Failed to validate request",
		},
		Fields: fields,
	}
}

// Predefined common apperror
var (
	ErrNotFound      = NewAppError(ErrCodeNotFound, "Resource not found")
	ErrUnauthorized  = NewAppError(ErrCodeUnauthorized, "Authentication required")
	ErrForbidden     = NewAppError(ErrCodeForbidden, "Access forbidden")
	ErrInternalError = NewAppError(ErrCodeInternalError, "Internal server error")
	ErrInvalidInput  = NewAppError(ErrCodeInvalidInput, "Invalid input provided")
	ErrConflict      = NewAppError(ErrCodeConflict, "Resource conflict")

	// Business logic apperror
	ErrUserAlreadyExists     = NewAppError(ErrCodeUserAlreadyExists, "User already exists")
	ErrUserNotFound          = NewAppError(ErrCodeUserNotFound, "User not found")
	ErrUserIncorrectPassword = NewAppError(ErrCodeUserIncorrectPassword, "Incorrect password")
	ErrUserInactive          = NewAppError(ErrCodeUserInactive, "User is inactive")
	ErrUserEmailNotVerified  = NewAppError(ErrCodeUserEmailNotVerified, "Email not verified")
	ErrInvalidOTP            = NewAppError(ErrCodeInvalidOTP, "Invalid OTP")
	ErrOTPExpired            = NewAppError(ErrCodeOTPExpired, "OTP has expired")
	ErrInvalidToken          = NewAppError(ErrCodeInvalidToken, "Invalid token")
)

// IsAppError checks if an error is an modules error
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) (*ValidationErrors, bool) {
	var validationErr *ValidationErrors
	if errors.As(err, &validationErr) {
		return validationErr, true
	}
	return nil, false
}

// WrapError wraps a standard error into an modules error
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return as is
	if appErr, ok := IsAppError(err); ok {
		return appErr
	}

	return NewAppErrorWithCause(code, message, err)
}

func NewBodyParserError() *AppError {
	return NewAppError(ErrCodeInvalidInput, "Invalid request body")
}
