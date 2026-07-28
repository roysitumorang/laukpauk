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
	initLogger = sync.OnceFunc(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelError,
		}))
	})
)

func NewLogger() *slog.Logger {
	initLogger()
	return logger
}

func logContext(args ...string) []any {
	n := len(args)
	fields := make([]any, 2+n/2)
	var j int
	fields[j] = slog.String("topic", Sprintf("%s-service-log", config.AppName))
	j++
	fields[j] = slog.String("service", config.AppName)
	j++
	for i := 0; i < n; i += 2 {
		if i+1 >= n {
			break
		}
		fields[j] = slog.String(args[i], args[i+1])
		j++
	}
	return fields
}

func Log(ctx context.Context, level slog.Level, message, context, scope string) {
	var args []string
	if context != "" {
		args = append(args, "context", context)
	}
	if scope != "" {
		args = append(args, "scope", scope)
	}
	entries := logContext(args...)
	switch level {
	case slog.LevelDebug:
		logger.DebugContext(ctx, message, entries...)
	case slog.LevelInfo:
		logger.InfoContext(ctx, message, entries...)
	case slog.LevelWarn:
		logger.WarnContext(ctx, message, entries...)
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
		logger.ErrorContext(ctx, message, entries...)
	}
}

func Capture(ctx context.Context, level slog.Level, err error, context, scope string) {
	var args []string
	if context != "" {
		args = append(args, "context", context)
	}
	if scope != "" {
		args = append(args, "scope", scope)
	}
	entries := logContext(args...)
	switch level {
	case slog.LevelDebug:
		logger.DebugContext(ctx, err.Error(), entries...)
	case slog.LevelInfo:
		logger.InfoContext(ctx, err.Error(), entries...)
	case slog.LevelWarn:
		logger.WarnContext(ctx, err.Error(), entries...)
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
		logger.ErrorContext(ctx, err.Error(), entries...)
	}
}
