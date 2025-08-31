package validator

import (
	"fmt"
)

// getEnglishErrorMessage returns English error messages
func (v *GoPlaygroundValidator) getEnglishErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", field)
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "contains":
		return fmt.Sprintf("%s must contain '%s'", field, param)
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of '%s'", field, param)
	case "excludes":
		return fmt.Sprintf("%s must not contain '%s'", field, param)
	case "startswith":
		return fmt.Sprintf("%s must start with '%s'", field, param)
	case "endswith":
		return fmt.Sprintf("%s must end with '%s'", field, param)
	case "uppercase":
		return fmt.Sprintf("%s must be uppercase", field)
	case "lowercase":
		return fmt.Sprintf("%s must be lowercase", field)
	case "base64":
		return fmt.Sprintf("%s must be a valid base64 string", field)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "hexadecimal":
		return fmt.Sprintf("%s must be a valid hexadecimal", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", field)
	case "rgb":
		return fmt.Sprintf("%s must be a valid RGB color", field)
	case "rgba":
		return fmt.Sprintf("%s must be a valid RGBA color", field)
	case "hsl":
		return fmt.Sprintf("%s must be a valid HSL color", field)
	case "hsla":
		return fmt.Sprintf("%s must be a valid HSLA color", field)
	case "e164":
		return fmt.Sprintf("%s must be a valid E.164 phone number", field)
	case "isbn":
		return fmt.Sprintf("%s must be a valid ISBN", field)
	case "isbn10":
		return fmt.Sprintf("%s must be a valid ISBN-10", field)
	case "isbn13":
		return fmt.Sprintf("%s must be a valid ISBN-13", field)
	case "credit_card":
		return fmt.Sprintf("%s must be a valid credit card number", field)
	case "ssn":
		return fmt.Sprintf("%s must be a valid SSN", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", field)
	case "password":
		return fmt.Sprintf("%s must contain at least 6 characters with uppercase, lowercase, number, and special character", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least 8 characters with 2+ uppercase, 2+ lowercase, 2+ numbers, 2+ special characters, and no common patterns", field)
	case "phone_id":
		return fmt.Sprintf("%s must be a valid Indonesian phone number", field)
	case "postal_code_id":
		return fmt.Sprintf("%s must be a valid Indonesian postal code (5 digits)", field)
	case "nik":
		return fmt.Sprintf("%s must be a valid Indonesian NIK (16 digits)", field)
	case "username":
		return fmt.Sprintf("%s must be a valid username (3-20 characters, alphanumeric, underscore, dot allowed)", field)
	case "no_html":
		return fmt.Sprintf("%s must not contain HTML tags", field)
	case "currency_id":
		return fmt.Sprintf("%s must be a valid Indonesian currency format", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// getIndonesianErrorMessage returns Indonesian error messages
func (v *GoPlaygroundValidator) getIndonesianErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s wajib diisi", field)
	case "email":
		return fmt.Sprintf("%s harus berupa alamat email yang valid", field)
	case "min":
		return fmt.Sprintf("%s minimal harus %s karakter", field, param)
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", field, param)
	case "len":
		return fmt.Sprintf("%s harus tepat %s karakter", field, param)
	case "numeric":
		return fmt.Sprintf("%s harus berupa angka", field)
	case "alpha":
		return fmt.Sprintf("%s hanya boleh mengandung huruf", field)
	case "alphanum":
		return fmt.Sprintf("%s hanya boleh mengandung huruf dan angka", field)
	case "url":
		return fmt.Sprintf("%s harus berupa URL yang valid", field)
	case "uri":
		return fmt.Sprintf("%s harus berupa URI yang valid", field)
	case "gte":
		return fmt.Sprintf("%s harus lebih besar atau sama dengan %s", field, param)
	case "lte":
		return fmt.Sprintf("%s harus lebih kecil atau sama dengan %s", field, param)
	case "gt":
		return fmt.Sprintf("%s harus lebih besar dari %s", field, param)
	case "lt":
		return fmt.Sprintf("%s harus lebih kecil dari %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s harus berupa UUID yang valid", field)
	case "datetime":
		return fmt.Sprintf("%s harus berupa tanggal dan waktu yang valid", field)
	case "eqfield":
		return fmt.Sprintf("%s harus sama dengan %s", field, param)
	case "nefield":
		return fmt.Sprintf("%s tidak boleh sama dengan %s", field, param)
	case "contains":
		return fmt.Sprintf("%s harus mengandung '%s'", field, param)
	case "containsany":
		return fmt.Sprintf("%s harus mengandung minimal salah satu dari '%s'", field, param)
	case "excludes":
		return fmt.Sprintf("%s tidak boleh mengandung '%s'", field, param)
	case "startswith":
		return fmt.Sprintf("%s harus dimulai dengan '%s'", field, param)
	case "endswith":
		return fmt.Sprintf("%s harus diakhiri dengan '%s'", field, param)
	case "uppercase":
		return fmt.Sprintf("%s harus berupa huruf kapital", field)
	case "lowercase":
		return fmt.Sprintf("%s harus berupa huruf kecil", field)
	case "base64":
		return fmt.Sprintf("%s harus berupa string base64 yang valid", field)
	case "json":
		return fmt.Sprintf("%s harus berupa JSON yang valid", field)
	case "hexadecimal":
		return fmt.Sprintf("%s harus berupa heksadesimal yang valid", field)
	case "hexcolor":
		return fmt.Sprintf("%s harus berupa warna hex yang valid", field)
	case "rgb":
		return fmt.Sprintf("%s harus berupa warna RGB yang valid", field)
	case "rgba":
		return fmt.Sprintf("%s harus berupa warna RGBA yang valid", field)
	case "hsl":
		return fmt.Sprintf("%s harus berupa warna HSL yang valid", field)
	case "hsla":
		return fmt.Sprintf("%s harus berupa warna HSLA yang valid", field)
	case "e164":
		return fmt.Sprintf("%s harus berupa nomor telepon E.164 yang valid", field)
	case "isbn":
		return fmt.Sprintf("%s harus berupa ISBN yang valid", field)
	case "isbn10":
		return fmt.Sprintf("%s harus berupa ISBN-10 yang valid", field)
	case "isbn13":
		return fmt.Sprintf("%s harus berupa ISBN-13 yang valid", field)
	case "credit_card":
		return fmt.Sprintf("%s harus berupa nomor kartu kredit yang valid", field)
	case "ssn":
		return fmt.Sprintf("%s harus berupa SSN yang valid", field)
	case "latitude":
		return fmt.Sprintf("%s harus berupa lintang yang valid", field)
	case "longitude":
		return fmt.Sprintf("%s harus berupa bujur yang valid", field)
	case "password":
		return fmt.Sprintf("%s harus mengandung minimal 6 karakter dengan huruf besar, huruf kecil, angka, dan karakter khusus", field)
	case "strong_password":
		return fmt.Sprintf("%s harus mengandung minimal 8 karakter dengan 2+ huruf besar, 2+ huruf kecil, 2+ angka, 2+ karakter khusus, dan tidak menggunakan pola umum", field)
	case "phone_id":
		return fmt.Sprintf("%s harus berupa nomor telepon Indonesia yang valid", field)
	case "postal_code_id":
		return fmt.Sprintf("%s harus berupa kode pos Indonesia yang valid (5 digit)", field)
	case "nik":
		return fmt.Sprintf("%s harus berupa NIK Indonesia yang valid (16 digit)", field)
	case "username":
		return fmt.Sprintf("%s harus berupa username yang valid (3-20 karakter, huruf, angka, underscore, titik diperbolehkan)", field)
	case "no_html":
		return fmt.Sprintf("%s tidak boleh mengandung tag HTML", field)
	case "currency_id":
		return fmt.Sprintf("%s harus berupa format mata uang Indonesia yang valid", field)
	default:
		return fmt.Sprintf("%s tidak valid", field)
	}
}
