// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tklauser/go-sysconf"
)

type contextKey string

const (
	dbusContextKey        contextKey = "dbus"
	clktckContextKey      contextKey = "clktck"
	boottimeContextKey    contextKey = "boottime"
	sessionPathContextKey contextKey = "sessionPath"
)

var ErrInvalidCtx = errors.New("invalid context")

func NewContext(ctx context.Context) (context.Context, error) {
	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	if err != nil {
		return nil, fmt.Errorf("cannot setup context: %w", err)
	}

	ctx = context.WithValue(ctx, clktckContextKey, clktck)

	boottime, err := getBootTime()
	if err != nil {
		return nil, fmt.Errorf("cannot setup context: %w", err)
	}

	ctx = context.WithValue(ctx, boottimeContextKey, boottime)

	return ctx, nil
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
