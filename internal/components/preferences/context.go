// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package preferences

import (
	"context"
	"path/filepath"

	"github.com/adrg/xdg"
)

type contextKey string

const (
	registrationCtxKey contextKey = "registration"
	pathCtxKey         contextKey = "path"
)

// RegistrationToCtx stores the registration details passed on the
// command-line to the context.
func RegistrationToCtx(ctx context.Context, registration Registration) context.Context {
	newCtx := context.WithValue(ctx, registrationCtxKey, registration)
	return newCtx
}

// RegistrationFromCtx retrieves the registration details passed on the
// command-line from the context.
func RegistrationFromCtx(ctx context.Context) *Registration {
	registration, ok := ctx.Value(registrationCtxKey).(Registration)
	if !ok {
		return nil
	}

	return &registration
}

// PathToCtx stores the base path for preferences (and other files) in the
// context.
func PathToCtx(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathCtxKey, path)
}

// PathFromCtx retrieves the base path for preferences from the context.
func PathFromCtx(ctx context.Context) string {
	path, ok := ctx.Value(pathCtxKey).(string)
	if !ok {
		return filepath.Join(xdg.ConfigHome, AppID)
	}

	return path
}
