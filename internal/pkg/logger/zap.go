package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger is a wrapper around zap.SugaredLogger
type ZapLogger struct {
	*zap.SugaredLogger
}

// NewZapLogger creates a new logger based on configuration
func NewZapLogger(cfg *config.Config) (*ZapLogger, error) {

	// This slice will hold all the configured logger cores (e.g., file, stdout)
	var cores []zapcore.Core

	// Get the configured log level and encoder
	logLevel := getLogLevel(cfg.LoggerLevel)
	encoder := getEncoder()

	// Parse the driver string, e.g., "file|stdout" -> ["file", "stdout"]
	drivers := strings.Split(cfg.LoggerOutput, "|")

	for _, driver := range drivers {
		driver = strings.TrimSpace(driver) // Clean up whitespace
		switch driver {
		case "file":
			// Configure lumberjack for log rotation
			lumberjackLogger := &lumberjack.Logger{
				Filename:   cfg.LoggerFile,
				MaxSize:    cfg.LoggerMaxSize,
				MaxBackups: cfg.LoggerMaxBackups,
				MaxAge:     cfg.LoggerMaxAge,
				Compress:   cfg.LoggerCompress,
				LocalTime:  true,
			}
			fileWriter := zapcore.AddSync(lumberjackLogger)
			// Create a zap core that writes to the file
			fileCore := zapcore.NewCore(encoder, fileWriter, logLevel)
			cores = append(cores, fileCore)
		case "stdout":
			// Create a zap core that writes to the console
			consoleWriter := zapcore.AddSync(os.Stdout)
			consoleCore := zapcore.NewCore(encoder, consoleWriter, logLevel)
			cores = append(cores, consoleCore)
		}
	}

	// If no valid drivers were found, return an error
	if len(cores) == 0 {
		return nil, fmt.Errorf("no valid logger drivers configured in '%s'. supported drivers: file, stdout", cfg.LoggerOutput)
	}

	// Combine all configured cores into one.
	// If only one driver is configured, it will use that one.
	// If multiple are configured, it will log to all of them.
	core := zapcore.NewTee(cores...)

	// Create the final logger with the combined core
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))

	return &ZapLogger{logger.Sugar()}, nil
}

// getEncoder returns a zap encoder
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogLevel returns a zap log level
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (z *ZapLogger) WithComponent(component string) Logger {
	return &ZapLogger{z.SugaredLogger.With("component", component)}
}

func (z *ZapLogger) WithContext(ctx context.Context) Logger {
	requestID, ok := ctx.Value("request_id").(string)
	if !ok {
		requestID = uuid.Must(uuid.NewV7()).String()
	}
	return &ZapLogger{z.SugaredLogger.With("request_id", requestID)}
}

func (z *ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Infow(msg, keysAndValues...)
}

func (z *ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (z *ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (z *ZapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Fatalw(msg, keysAndValues...)
}

func (z *ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Debugw(msg, keysAndValues...)
}

func (z *ZapLogger) With(keysAndValues ...interface{}) Logger {
	return &ZapLogger{z.SugaredLogger.With(keysAndValues...)}
}
