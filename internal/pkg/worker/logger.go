package worker

import (
	"context"
	"fmt"

	"github.com/sammidev/goca/internal/pkg/logger"
)

type Logger struct {
	logger logger.Logger
}

func NewLogger(logger logger.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *Logger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *Logger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *Logger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *Logger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}
