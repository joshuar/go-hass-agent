// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

// setupWorkerCtx sets up the worker context for Linux systems.
func setupWorkerCtx(ctx context.Context) context.Context {
	return linux.NewContext(ctx)
}
