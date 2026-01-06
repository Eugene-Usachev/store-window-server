package logger

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	inner zerolog.Logger
}

var _ log.Logger = (*ZeroLogger)(nil)

func NewZeroLogger() log.Logger {
	return &ZeroLogger{
		inner: zerolog.New(zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: true,
		}).With().Timestamp().Logger(),
	}
}

func (z *ZeroLogger) Log(level log.Level, keyvals ...any) error {
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "(missing)")
	}

	var e *zerolog.Event

	switch level {
	case log.LevelDebug:
		e = z.inner.Debug()
	case log.LevelInfo:
		e = z.inner.Info()
	case log.LevelWarn:
		e = z.inner.Warn()
	case log.LevelError:
		e = z.inner.Error()
	case log.LevelFatal:
		e = z.inner.Fatal()
	default:
		e = z.inner.Log()
	}

	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}

		e = e.Interface(key, keyvals[i+1])
	}

	e.Send()

	return nil
}
