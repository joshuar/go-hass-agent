// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

// setupWorkerCtx sets up the worker context for Linux systems. This is a
// wrapper around the OS-specific method to set up a context and is only used on
// Linux systems (via file filtering when building).
func setupWorkerCtx(ctx context.Context) context.Context {
	return linux.NewContext(ctx)
}
