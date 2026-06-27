package helper

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/config"
)

var (
	logger     *slog.Logger
	InitLogger = sync.OnceFunc(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
		}))
	})
)

func GetLogger() *slog.Logger {
	return logger
}

func logContext(context, scope string) []any {
	fields := []any{
		slog.String("topic", Sprintf("%s-service-log", config.AppName)),
		slog.String("context", context),
		slog.String("service", config.AppName),
	}
	if scope != "" {
		fields = append(fields, slog.String("scope", scope))
	}
	return fields
}

func Log(ctx context.Context, level slog.Level, message, context, scope string) {
	entries := logContext(context, scope)
	switch level {
	case slog.LevelDebug:
		logger.Debug(message, entries...)
	case slog.LevelInfo:
		logger.Info(message, entries...)
	case slog.LevelWarn:
		logger.Warn(message, entries...)
	case slog.LevelError:
		var name string
		pc, file, line, _ := runtime.Caller(1)
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
		entries = append(
			entries,
			slog.String("func", name),
			slog.String("file", Sprintf("%s:%d", file, line)),
			slog.Int("line", line),
		)
		logger.Error(message, entries...)
	}
}

func Capture(ctx context.Context, level slog.Level, err error, context, scope string) {
	entries := logContext(context, scope)
	switch level {
	case slog.LevelDebug:
		logger.Debug(err.Error(), entries...)
	case slog.LevelInfo:
		logger.Info(err.Error(), entries...)
	case slog.LevelWarn:
		logger.Warn(err.Error(), entries...)
	case slog.LevelError:
		// ignoring pgx.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		var name string
		pc, file, line, _ := runtime.Caller(1)
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
		entries = append(
			entries,
			slog.String("func", name),
			slog.String("file", Sprintf("%s:%d", file, line)),
			slog.Int("line", line),
		)
		logger.Error(err.Error(), entries...)
	}
}
