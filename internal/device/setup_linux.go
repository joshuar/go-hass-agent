// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

// SetupCtx sets up a context for Linux systems. This is a wrapper around the
// OS-specific method to set up a context and is only used on Linux systems (via
// file filtering when building).
func SetupCtx(ctx context.Context) context.Context {
	return linux.NewContext(ctx)
}
