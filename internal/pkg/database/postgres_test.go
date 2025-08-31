package database

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/sammidev/goca/internal/pkg/logger"
	. "github.com/smartystreets/goconvey/convey"
)

// mockLogger adalah implementasi palsu dari logger.Logger untuk pengujian.
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

func TestPostgreSQLDatabase(t *testing.T) {
	Convey("Diberikan instance PostgreSQLDatabase dengan mock pool", t, func() {
		mockPool, err := pgxmock.NewPool()
		So(err, ShouldBeNil)
		defer mockPool.Close()

		db := &PostgreSQLDatabase{
			pool:   mockPool,
			logger: newMockLogger(),
		}
		ctx := context.Background()

		Convey("Saat memanggil GetSQLExecutor tanpa transaksi", func() {
			executor, err := db.GetSQLExecutor(ctx)
			So(err, ShouldBeNil)

			Convey("Executor yang dikembalikan seharusnya adalah pool", func() {
				So(executor, ShouldEqual, mockPool)
			})
		})

		Convey("Saat memanggil Close", func() {
			// pgxmock tidak mengekspos ekspektasi untuk Close(), tetapi kita dapat memverifikasi
			// bahwa tidak ada kesalahan yang terjadi.
			db.Close()
		})
	})

	Convey("Fungsi pembantu", t, func() {
		Convey("IsUniqueViolation", func() {
			So(IsUniqueViolation(errors.New("bebas")), ShouldBeFalse)
			So(IsUniqueViolation(errors.New("error: duplicate key value violates unique constraint \"users_email_key\"")), ShouldBeTrue)
			So(IsUniqueViolation(errors.New("ERROR: unique_violation (SQLSTATE 23505)")), ShouldBeTrue)
			So(IsUniqueViolation(nil), ShouldBeFalse)
		})
	})
}
