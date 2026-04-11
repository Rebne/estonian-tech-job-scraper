package logger

import (
	"context"
	"log/slog"
)

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default() // fallback
}
