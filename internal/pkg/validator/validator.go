package validator

// Validator defines the interface for validation operations
type Validator interface {
	Validate(i any) error
	ValidateAndGetErrors(i any) ValidationErrors
	SetLocale(locale Locale)
	GetLocale() Locale
}
