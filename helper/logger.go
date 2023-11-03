package helper

import (
	"context"
	"errors"
	"runtime"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Topic = "laukpauk-service-log"
)

var (
	service = "laukpauk"
)

func logContext(_ context.Context, context, scope string) *zap.Logger {
	logger := zap.Must(zap.NewProduction())
	defer func() {
		_ = logger.Sync()
	}()
	childLogger := logger.With(
		zap.String("topic", Topic),
		zap.String("context", context),
		zap.String("scope", scope),
		zap.String("service", service),
	)
	return childLogger
}

func Log(ctx context.Context, level zapcore.Level, message, context, scope string) {
	entry := logContext(ctx, context, scope)
	switch level {
	case zap.DebugLevel:
		entry.Debug(message)
	case zap.InfoLevel:
		entry.Info(message)
	case zap.WarnLevel:
		entry.Warn(message)
	case zap.ErrorLevel:
		entry.Error(message)
	case zap.FatalLevel:
		entry.Fatal(message)
	case zap.PanicLevel:
		entry.Panic(message)
	}
}

func Capture(ctx context.Context, level zapcore.Level, err error, context, scope string) {
	entry := logContext(ctx, context, scope)
	switch level {
	case zap.DebugLevel:
		entry.Debug(err.Error())
	case zap.InfoLevel:
		entry.Info(err.Error())
	case zap.WarnLevel:
		entry.Warn(err.Error())
	case zap.ErrorLevel:
		// ignoring pgx.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		var (
			name   string
			pgxErr *pgconn.PgError
		)
		pc, file, line, _ := runtime.Caller(1)
		if !errors.As(err, &pgxErr) {
			pc, file, line, _ = runtime.Caller(4)
		}
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
		entry.Error(
			err.Error(),
			zap.String("func", name),
			zap.String("file", file),
			zap.Int("line", line),
		)
	case zap.FatalLevel:
		entry.Fatal(err.Error())
	case zap.PanicLevel:
		entry.Panic(err.Error())
	}
}
