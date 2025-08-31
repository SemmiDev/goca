package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

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

func TestGoCronScheduler(t *testing.T) {
	Convey("Diberikan Scheduler gocron", t, func() {
		mockLog := newMockLogger()

		Convey("Fungsi New", func() {
			Convey("Seharusnya berhasil membuat scheduler baru", func() {
				s, err := New(mockLog)
				So(err, ShouldBeNil)
				So(s, ShouldNotBeNil)
			})
		})

		Convey("Metode Scheduler", func() {
			s, err := New(mockLog)
			So(err, ShouldBeNil)

			Convey("RegisterJob", func() {
				Convey("Seharusnya berhasil mendaftarkan job dengan jadwal yang valid", func() {
					err := s.RegisterJob("1s", func() {})
					So(err, ShouldBeNil)
				})

				Convey("Seharusnya gagal mendaftarkan job dengan jadwal yang tidak valid", func() {
					err := s.RegisterJob("jadwal tidak valid", func() {})
					So(err, ShouldNotBeNil)
				})
			})

			Convey("Start dan Stop", func() {
				Convey("Seharusnya menjalankan job yang terdaftar setelah dimulai", func() {
					var wg sync.WaitGroup
					wg.Add(1)

					jobHasRun := false
					jobFunc := func() {
						jobHasRun = true
						wg.Done()
					}

					err := s.RegisterJob("1s", jobFunc)
					So(err, ShouldBeNil)

					s.Start()

					// Tunggu job selesai atau timeout setelah beberapa detik
					c := make(chan struct{})
					go func() {
						defer close(c)
						wg.Wait()
					}()

					select {
					case <-c:
						// Job selesai dengan sukses
					case <-time.After(2 * time.Second):
						t.Fatal("Timeout menunggu job scheduler berjalan")
					}

					s.Stop()

					So(jobHasRun, ShouldBeTrue)
				})
			})
		})
	})
}
