// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/tklauser/go-sysconf"

	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type contextKey string

const (
	dbusSessionContextKey   contextKey = "sessionBus"
	dbusSystemContextKey    contextKey = "systemBus"
	clktckContextKey        contextKey = "clktck"
	boottimeContextKey      contextKey = "boottime"
	sessionPathContextKey   contextKey = "sessionPath"
	desktopPortalContextKey contextKey = "desktopPortal"
)

var (
	ErrInvalidCtx      = errors.New("invalid context")
	ErrNoSessionBus    = errors.New("no session bus connection in context")
	ErrNoSystemBus     = errors.New("no system bus connection in context")
	ErrNoDesktopPortal = errors.New("no desktop portal in context")
	ErrNoSessionPath   = errors.New("no session path in context")
)

func NewContext(ctx context.Context) context.Context {
	// Add clock ticks value.
	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	if err != nil {
		slog.Warn("Unable to add system clock ticks to context.", slog.Any("error", err))
	}

	ctx = context.WithValue(ctx, clktckContextKey, clktck)

	// Add boot time value.
	boottime, err := getBootTime()
	if err != nil {
		slog.Warn("Unable to add boot time to context.", slog.Any("error", err))
	}

	ctx = context.WithValue(ctx, boottimeContextKey, boottime)

	// Add portal interface
	portal, err := FindPortal()
	if err != nil {
		slog.Warn("Unable to add desktop portal to context.", slog.Any("error", err))
	}

	ctx = context.WithValue(ctx, desktopPortalContextKey, portal)

	// Add D-Bus system bus connection.
	systemBus, err := dbusx.NewBus(ctx, dbusx.SystemBus)
	if err != nil {
		slog.Warn("Unable to set up D-Bus system bus connection.", slog.Any("error", err))
	}

	ctx = context.WithValue(ctx, dbusSystemContextKey, systemBus)

	// Add session path value.
	sessionPath, err := systemBus.GetSessionPath()
	if err != nil {
		slog.Warn("Unable to determine user session path from D-Bus.", slog.Any("error", err))
	}

	ctx = context.WithValue(ctx, sessionPathContextKey, sessionPath)

	// Add D-Bus session bus connection.
	sessionBus, err := dbusx.NewBus(ctx, dbusx.SessionBus)
	if err != nil {
		slog.Warn("Unable to set up D-Bus session bus connection.", slog.Any("error", err))
	}

	ctx = context.WithValue(ctx, dbusSessionContextKey, sessionBus)

	return ctx
}

func CtxGetClkTck(ctx context.Context) (int64, bool) {
	clktck, ok := ctx.Value(clktckContextKey).(int64)
	if !ok {
		return 0, false
	}

	return clktck, true
}

func CtxGetBoottime(ctx context.Context) (time.Time, bool) {
	boottime, ok := ctx.Value(boottimeContextKey).(time.Time)
	if !ok {
		return time.Now(), false
	}

	return boottime, true
}

func CtxGetDesktopPortal(ctx context.Context) (string, bool) {
	portal, ok := ctx.Value(desktopPortalContextKey).(string)
	if !ok {
		return portal, false
	}

	return portal, true
}

func CtxGetSystemBus(ctx context.Context) (*dbusx.Bus, bool) {
	bus, ok := ctx.Value(dbusSystemContextKey).(*dbusx.Bus)
	if !ok {
		return bus, false
	}

	return bus, true
}

func CtxGetSessionBus(ctx context.Context) (*dbusx.Bus, bool) {
	bus, ok := ctx.Value(dbusSessionContextKey).(*dbusx.Bus)
	if !ok {
		return bus, false
	}

	return bus, true
}

func CtxGetSessionPath(ctx context.Context) (string, bool) {
	path, ok := ctx.Value(sessionPathContextKey).(string)
	if !ok {
		return path, false
	}

	return path, true
}
