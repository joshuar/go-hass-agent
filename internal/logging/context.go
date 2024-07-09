// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	loggerContextKey contextKey = "logger"
)

func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	newCtx := context.WithValue(ctx, loggerContextKey, logger)

	return newCtx
}

func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return logger
}
