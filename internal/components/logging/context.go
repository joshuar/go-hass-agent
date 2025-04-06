// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package logging

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	loggerContextKey contextKey = "logger"
)

// ToContext will store the given logger in the context.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	newCtx := context.WithValue(ctx, loggerContextKey, logger)

	return newCtx
}

// FromContext will retrieve a logger from the context. If there is no logger in the context, the default logger is
// returned instead.
func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return logger
}
