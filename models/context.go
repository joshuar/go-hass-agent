// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import "context"

const (
	csrfTokenCtxKey contextKey = "csrfToken"
)

type contextKey string

func CSRFTokenToCtx(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfTokenCtxKey, token)
}

func CSRFTokenFromCtx(ctx context.Context) string {
	if token, ok := ctx.Value(csrfTokenCtxKey).(string); ok {
		return token
	}
	return ""
}
