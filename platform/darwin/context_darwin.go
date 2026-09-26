package darwin

import (
	"context"
	"log/slog"

	"github.com/tklauser/go-sysconf"
)

type contextKey string

const (
	clktckContextKey   contextKey = "clktck"
	boottimeContextKey contextKey = "boottime"
)

func NewContext(ctx context.Context) context.Context {
	// Add clock ticks value.
	if clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK); err != nil {
		slog.Warn("Unable to add system clock ticks to context. Some sensors requring it may not be available",
			slog.Any("error", err),
		)
	} else {
		ctx = context.WithValue(ctx, clktckContextKey, clktck)
	}

	// Add boot time value.
	if boottime, err := getBootTime(); err != nil {
		slog.Warn("Unable to add boot time to context. Some sensors requring it may not be available",
			slog.Any("error", err),
		)
	} else {
		ctx = context.WithValue(ctx, boottimeContextKey, boottime)
	}

	return ctx
}
