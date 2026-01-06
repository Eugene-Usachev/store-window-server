package logger

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

type Logger struct {
	log.Helper
}

func InitMainLogger() Logger {
	logger := Logger{
		*log.NewHelper(log.With(
			NewZeroLogger(),
			"ts", log.DefaultTimestamp,
			"caller", log.DefaultCaller,
			"trace.id", tracing.TraceID(),
			"span.id", tracing.SpanID(),
		)),
	}

	log.SetLogger(logger.Logger())

	return logger
}

var mainLoggerInstance = InitMainLogger()

func MainLogger() *Logger {
	return &mainLoggerInstance
}

func (l *Logger) Print(msg string) {
	l.Info(msg)
}

func (l *Logger) Printf(msg string, args ...any) {
	l.Infof(msg, args...)
}

func Info(msg string) {
	MainLogger().Info(msg)
}

func Infof(msg string, args ...any) {
	MainLogger().Infof(msg, args...)
}

func Error(msg string) {
	MainLogger().Error(msg)
}

func Errorf(msg string, args ...any) {
	MainLogger().Errorf(msg, args...)
}

func Fatal(msg string) {
	MainLogger().Fatal(msg)
}

func Fatalf(msg string, args ...any) {
	MainLogger().Fatalf(msg, args...)
}
