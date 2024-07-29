// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"context"
	"errors"
)

type contextKey string

const (
	restAPIURLContextKey   contextKey = "restAPIURL"
	websocketURLContextKey contextKey = "websocketURL"
	tokenContextKey        contextKey = "token"
	webhookIDContextKey    contextKey = "webhookid"
)

var (
	ErrNoRestAPIURLInCtx   = errors.New("no rest API URL in context")
	ErrNoWebSocketURLInCtx = errors.New("no web socket URL in context")
	ErrNoTokenInCtx        = errors.New("no token in context")
	ErrNoWebhookIDInCtx    = errors.New("no webhook id in context")
)

// ContextSetRestAPIURL will store the preferences in the given context.
func ContextSetRestAPIURL(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, restAPIURLContextKey, url)
}

// ContextGetRestAPIURL will attempt to fetch the preferences from the given context.
func ContextGetRestAPIURL(ctx context.Context) (string, error) {
	prefs, ok := ctx.Value(restAPIURLContextKey).(string)
	if !ok {
		return "", ErrNoRestAPIURLInCtx
	}

	return prefs, nil
}

// ContextSetRestAPIURL will store the preferences in the given context.
func ContextSetWebsocketURL(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, websocketURLContextKey, url)
}

// ContextGetRestAPIURL will attempt to fetch the preferences from the given context.
func ContextGetWebsocketURL(ctx context.Context) (string, error) {
	prefs, ok := ctx.Value(websocketURLContextKey).(string)
	if !ok {
		return "", ErrNoRestAPIURLInCtx
	}

	return prefs, nil
}

// ContextSetToken will store the preferences in the given context.
func ContextSetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

// ContextGetToken will attempt to fetch the preferences from the given context.
func ContextGetToken(ctx context.Context) (string, error) {
	prefs, ok := ctx.Value(tokenContextKey).(string)
	if !ok {
		return "", ErrNoTokenInCtx
	}

	return prefs, nil
}

// ContextSetToken will store the preferences in the given context.
func ContextSetWebhookID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, webhookIDContextKey, id)
}

// ContextGetToken will attempt to fetch the preferences from the given context.
func ContextGetWebhookID(ctx context.Context) (string, error) {
	prefs, ok := ctx.Value(webhookIDContextKey).(string)
	if !ok {
		return "", ErrNoWebhookIDInCtx
	}

	return prefs, nil
}
