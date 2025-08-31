package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Locale constants
type Locale string

const (
	LocaleEN Locale = "en" // English
	LocaleID Locale = "id" // Indonesian
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents a collection of validation apperror
type ValidationErrors []ValidationError

// ToMap converts validation apperror to a map for easier access
func (ve ValidationErrors) ToMap() map[string]string {
	m := make(map[string]string, len(ve))
	for _, err := range ve {
		m[err.Field] = err.Message
	}
	return m
}

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// GoPlaygroundValidator is a concrete implementation of the Validator interface using go-playground/validator
type GoPlaygroundValidator struct {
	validate *validator.Validate
	locale   Locale
}

// New creates a new GoPlaygroundValidator instance with English as the default locale
func New() Validator {
	return NewGoPlaygroundValidatorWithLocale(LocaleID)
}

// NewGoPlaygroundValidatorWithLocale creates a new GoPlaygroundValidator instance with the specified locale
func NewGoPlaygroundValidatorWithLocale(locale Locale) *GoPlaygroundValidator {
	validate := validator.New()

	// Register a custom tag name function to use "json" tags for field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	v := &GoPlaygroundValidator{
		validate: validate,
		locale:   locale,
	}

	// Register custom validators
	v.registerCustomValidators()

	return v
}

// SetLocale changes the validator's locale
func (v *GoPlaygroundValidator) SetLocale(locale Locale) {
	v.locale = locale
}

// GetLocale returns the current locale
func (v *GoPlaygroundValidator) GetLocale() Locale {
	return v.locale
}

// Validate validates the input struct and returns any validation apperror
func (v *GoPlaygroundValidator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	var validationErrors ValidationErrors
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			validationError := ValidationError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
				Message: v.getErrorMessage(fieldErr),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// ValidateAndGetErrors validates the input struct and returns the validation apperror directly
func (v *GoPlaygroundValidator) ValidateAndGetErrors(i any) ValidationErrors {
	err := v.Validate(i)
	if err == nil {
		return nil
	}

	if validationErrs, ok := err.(ValidationErrors); ok {
		return validationErrs
	}

	return nil
}

// getErrorMessage generates localized error messages based on validation tag and locale
func (v *GoPlaygroundValidator) getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	param := fe.Param()
	tag := fe.Tag()

	switch v.locale {
	case LocaleID:
		return v.getIndonesianErrorMessage(field, tag, param)
	default:
		return v.getEnglishErrorMessage(field, tag, param)
	}
}

// Helper functions for convenience

// IsValidationErrors checks if the error is of type ValidationErrors
func IsValidationErrors(err error) bool {
	_, ok := err.(ValidationErrors)
	return ok
}

// GetValidationErrors safely converts error to ValidationErrors
func GetValidationErrors(err error) (ValidationErrors, bool) {
	if validationErrs, ok := err.(ValidationErrors); ok {
		return validationErrs, true
	}
	return nil, false
}

// registerCustomValidators registers all custom validation tags
func (v *GoPlaygroundValidator) registerCustomValidators() {
	// Password validation with comprehensive rules
	v.validate.RegisterValidation("password", v.validatePassword)
	// Strong password with stricter rules
	v.validate.RegisterValidation("strong_password", v.validateStrongPassword)
	// Indonesian phone number
	v.validate.RegisterValidation("phone_id", v.validateIndonesianPhone)
	// Indonesian postal code
	v.validate.RegisterValidation("postal_code_id", v.validateIndonesianPostalCode)
	// Indonesian NIK
	v.validate.RegisterValidation("nik", v.validateIndonesianNIK)
	// Username
	v.validate.RegisterValidation("username", v.validateUsername)
	// No HTML tags
	v.validate.RegisterValidation("no_html", v.validateNoHTML)
	// Indonesian currency format
	v.validate.RegisterValidation("currency_id", v.validateIndonesianCurrency)
}

// validatePassword validates password with standard rules:
// - Minimum 6 characters
// - At least 1 lowercase letter
// - At least 1 uppercase letter
// - At least 1 digit
// - At least 1 special character
func (v *GoPlaygroundValidator) validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Minimum length
	if len(password) < 6 {
		return false
	}

	var (
		hasLower   = false
		hasUpper   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasNumber && hasSpecial
}

// validateStrongPassword validates password with stricter rules:
// - Minimum 8 characters
// - At least 2 lowercase letters
// - At least 2 uppercase letters
// - At least 2 digits
// - At least 2 special characters
// - No common patterns or sequences
func (v *GoPlaygroundValidator) validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Minimum length
	if len(password) < 8 {
		return false
	}

	var (
		lowerCount   = 0
		upperCount   = 0
		numberCount  = 0
		specialCount = 0
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			lowerCount++
		case unicode.IsUpper(char):
			upperCount++
		case unicode.IsNumber(char):
			numberCount++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			specialCount++
		}
	}

	// Check minimum counts
	if lowerCount < 2 || upperCount < 2 || numberCount < 2 || specialCount < 2 {
		return false
	}

	// Check for common weak patterns
	lowerPass := strings.ToLower(password)
	weakPatterns := []string{
		"123456", "654321", "abcdef", "fedcba",
		"qwerty", "asdfgh", "zxcvbn", "password",
		"admin", "user", "guest", "test",
	}

	for _, pattern := range weakPatterns {
		if strings.Contains(lowerPass, pattern) {
			return false
		}
	}

	return true
}

// validateIndonesianPhone validates an Indonesian phone number (e.g., +6281234567890 or 081234567890)
func (v *GoPlaygroundValidator) validateIndonesianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Indonesian phone numbers typically start with +62 or 08, followed by 8-12 digits
	re := regexp.MustCompile(`^(?:\+628|08)\d{8,12}$`)
	return re.MatchString(phone)
}

// validateIndonesianPostalCode validates an Indonesian postal code (5 digits)
func (v *GoPlaygroundValidator) validateIndonesianPostalCode(fl validator.FieldLevel) bool {
	postalCode := fl.Field().String()
	re := regexp.MustCompile(`^\d{5}$`)
	return re.MatchString(postalCode)
}

// validateIndonesianNIK validates an Indonesian NIK (16 digits)
func (v *GoPlaygroundValidator) validateIndonesianNIK(fl validator.FieldLevel) bool {
	nik := fl.Field().String()
	re := regexp.MustCompile(`^\d{16}$`)
	return re.MatchString(nik)
}

// validateUsername validates a username (3-20 characters, alphanumeric, underscore, dot allowed)
func (v *GoPlaygroundValidator) validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	re := regexp.MustCompile(`^[a-zA-Z0-9_.]{3,20}$`)
	return re.MatchString(username)
}

// validateNoHTML ensures the input does not contain HTML tags
func (v *GoPlaygroundValidator) validateNoHTML(fl validator.FieldLevel) bool {
	input := fl.Field().String()
	re := regexp.MustCompile(`<[^>]+>`)
	return !re.MatchString(input)
}

// validateIndonesianCurrency validates Indonesian currency format (e.g., Rp1.000.000 or Rp 1.000.000)
func (v *GoPlaygroundValidator) validateIndonesianCurrency(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	// Allow "Rp" or "Rp " followed by digits with optional thousand separators (.)
	re := regexp.MustCompile(`^Rp\s?\d{1,3}(\.\d{3})*(\,\d{2})?$`)
	return re.MatchString(currency)
}
