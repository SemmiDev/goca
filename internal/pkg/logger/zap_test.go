package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sammidev/goca/internal/config"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// helper function to create a test logger that writes to an in-memory buffer
func newTestLogger(level zapcore.Level) (*ZapLogger, *bytes.Buffer) {
	var buffer bytes.Buffer
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buffer), level)
	logger := zap.New(core).Sugar()
	return &ZapLogger{logger}, &buffer
}

func TestNewZapLogger(t *testing.T) {
	Convey("Diberikan fungsi NewZapLogger", t, func() {
		Convey("Ketika konfigurasi valid diberikan", func() {
			// Buat file log sementara
			tmpFile, err := os.CreateTemp("", "test-log-*.log")
			So(err, ShouldBeNil)
			filePath := tmpFile.Name()
			tmpFile.Close()
			// hapus file setelah test selesai
			defer os.Remove(filePath)

			cfg := &config.Config{
				LoggerFile:   filePath,
				LoggerLevel:  "debug",
				LoggerOutput: "file|stdout",
			}

			logger, err := NewZapLogger(cfg)

			Convey("Seharusnya membuat logger dengan sukses tanpa error", func() {
				So(err, ShouldBeNil)
				So(logger, ShouldNotBeNil)
			})
		})
	})
}

func TestGetLogLevel(t *testing.T) {
	Convey("Diberikan fungsi getLogLevel", t, func() {
		testCases := map[string]zapcore.Level{
			"debug":   zapcore.DebugLevel,
			"info":    zapcore.InfoLevel,
			"warn":    zapcore.WarnLevel,
			"error":   zapcore.ErrorLevel,
			"fatal":   zapcore.FatalLevel,
			"unknown": zapcore.InfoLevel, // Default case
		}

		for levelStr, expectedLevel := range testCases {
			Convey(`Ketika level adalah "`+levelStr+`"`, func() {
				level := getLogLevel(levelStr)
				Convey("Seharusnya mengembalikan level zap yang benar", func() {
					So(level, ShouldEqual, expectedLevel)
				})
			})
		}
	})
}

func TestZapLoggerMethods(t *testing.T) {
	Convey("Diberikan instance ZapLogger", t, func() {
		logger, buffer := newTestLogger(zapcore.DebugLevel)

		Convey("Metode WithComponent", func() {
			componentLogger := logger.WithComponent("TestComponent")
			componentLogger.Info("Pesan test komponen")

			var output map[string]interface{}
			err := json.Unmarshal(buffer.Bytes(), &output)
			So(err, ShouldBeNil)

			Convey("Seharusnya menambahkan field 'component' ke log", func() {
				So(output["component"], ShouldEqual, "TestComponent")
			})
		})

		Convey("Metode WithContext", func() {
			Convey("Ketika konteks memiliki request_id", func() {
				ctx := context.WithValue(context.Background(), "request_id", "test-req-123")
				contextLogger := logger.WithContext(ctx)
				contextLogger.Info("Pesan test konteks")

				var output map[string]interface{}
				err := json.Unmarshal(buffer.Bytes(), &output)
				So(err, ShouldBeNil)

				Convey("Seharusnya menambahkan 'request_id' dari konteks", func() {
					So(output["request_id"], ShouldEqual, "test-req-123")
				})
			})

			Convey("Ketika konteks tidak memiliki request_id", func() {
				contextLogger := logger.WithContext(context.Background())
				contextLogger.Info("Pesan test konteks tanpa id")

				var output map[string]interface{}
				err := json.Unmarshal(buffer.Bytes(), &output)
				So(err, ShouldBeNil)

				Convey("Seharusnya menghasilkan dan menambahkan 'request_id' baru", func() {
					So(output["request_id"], ShouldNotBeEmpty)
				})
			})
		})

		Convey("Metode logging (Info)", func() {
			logger.Info("Pesan info", "key1", "value1", "key2", 42)

			var output map[string]interface{}
			err := json.Unmarshal(buffer.Bytes(), &output)
			So(err, ShouldBeNil)

			Convey("Seharusnya mencatat pesan, level, dan field dengan benar", func() {
				So(output["level"], ShouldEqual, "INFO")
				So(output["msg"], ShouldEqual, "Pesan info")
				So(output["key1"], ShouldEqual, "value1")
				So(output["key2"], ShouldEqual, 42)
			})
		})

		Convey("Metode With", func() {
			withLogger := logger.With("user_id", "user-456", "tenant", "acme")
			withLogger.Warn("Pesan peringatan")

			var output map[string]interface{}
			err := json.Unmarshal(buffer.Bytes(), &output)
			So(err, ShouldBeNil)

			Convey("Seharusnya menambahkan semua field yang diberikan ke log", func() {
				So(output["level"], ShouldEqual, "WARN")
				So(output["msg"], ShouldEqual, "Pesan peringatan")
				So(output["user_id"], ShouldEqual, "user-456")
				So(output["tenant"], ShouldEqual, "acme")
			})
		})
	})
}
