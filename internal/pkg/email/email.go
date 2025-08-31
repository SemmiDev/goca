package email

type Email interface {
	Send(recipient, subject string, htmlContent string, data any) error
}
