package assets

import (
	"embed"
)

//go:embed "emails"
var EmbeddedFiles embed.FS

const (
	EmailVerificationTemplatePath   = "emails/email-verification.tmpl"
	EmailForgotPasswordTemplatePath = "emails/email-forgot-password.tmpl"
)
