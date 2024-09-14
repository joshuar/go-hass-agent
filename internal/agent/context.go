// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type contextKey string

const (
	headlessCtxKey      contextKey = "headless"
	forceRegisterCtxKey contextKey = "forceregister"
	ignoreURLsCtxKey    contextKey = "ignoreURLS"
	serverCtxKey        contextKey = "server"
	tokenCtxKey         contextKey = "token"
)

func addToContext[T any](ctx context.Context, key contextKey, value T) context.Context {
	newCtx := context.WithValue(ctx, key, value)

	return newCtx
}

func Headless(ctx context.Context) bool {
	headless, ok := ctx.Value(headlessCtxKey).(bool)
	if !ok {
		return false
	}

	return headless
}

func ForceRegister(ctx context.Context) bool {
	forceregister, ok := ctx.Value(forceRegisterCtxKey).(bool)
	if !ok {
		return false
	}

	return forceregister
}

func IgnoreURLs(ctx context.Context) bool {
	ignoreURLs, ok := ctx.Value(ignoreURLsCtxKey).(bool)
	if !ok {
		return false
	}

	return ignoreURLs
}

func Server(ctx context.Context) string {
	server, ok := ctx.Value(serverCtxKey).(string)
	if !ok {
		return preferences.DefaultServer
	}

	return server
}

func Token(ctx context.Context) string {
	token, ok := ctx.Value(tokenCtxKey).(string)
	if !ok {
		return preferences.DefaultSecret
	}

	return token
}
