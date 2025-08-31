package logger

import "context"

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	WithComponent(component string) Logger
	WithContext(ctx context.Context) Logger
	With(keysAndValues ...interface{}) Logger
}
