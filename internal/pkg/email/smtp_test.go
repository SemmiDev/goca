package email

import (
	"context"
	"testing"

	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/logger"
	. "github.com/smartystreets/goconvey/convey"
)

// mockLogger is a fake implementation of logger.Logger for testing purposes.
type mockLogger struct{}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{})   {}
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})   {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Debug(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Fatal(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) WithComponent(component string) logger.Logger    { return m }
func (m *mockLogger) WithContext(ctx context.Context) logger.Logger   { return m }
func (m *mockLogger) With(keysAndValues ...interface{}) logger.Logger { return m }

func newMockLogger() logger.Logger {
	return &mockLogger{}
}

func TestSmtp(t *testing.T) {
	Convey("Diberikan service SMTP", t, func() {
		mockLog := newMockLogger()
		cfg := &config.Config{
			SMTPHost:     "localhost",
			SMTPPort:     1025, // Port umum untuk server SMTP tes seperti MailHog
			SMTPUser:     "test@example.com",
			SMTPPassword: "password",
			AppName:      "Test App",
		}

		Convey("Fungsi New", func() {
			Convey("Seharusnya berhasil membuat instance SMTP baru dengan konfigurasi yang valid", func() {
				s, err := NewSMTPClient(cfg, mockLog)
				So(err, ShouldBeNil)
				So(s, ShouldNotBeNil)
			})
		})

		Convey("Metode Send", func() {
			s, err := NewSMTPClient(cfg, mockLog)
			So(err, ShouldBeNil)

			Convey("Seharusnya mengembalikan error jika gagal terhubung ke server SMTP", func() {
				// Konfigurasi ini dijamin gagal karena tidak ada server SMTP yang berjalan di port ini
				cfg.SMTPPort = 9999
				s, _ = NewSMTPClient(cfg, mockLog)

				recipient := "recipient@example.com"
				subject := "Subject Test"
				htmlContent := "<h1>Hello World</h1>"

				err := s.Send(recipient, subject, htmlContent, nil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
