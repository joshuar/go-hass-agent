// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package upgrade provides some limited assistance for automatic major version upgrades.
package upgrade

import (
	"context"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
)

// Run will run the upgrade command.
func Run(ctx context.Context) error {
	logging.FromContext(ctx).Info("Checking for and attempting pre v10.0.0 upgrades...")
	// Perform pre v10.0.0 upgrades...
	if err := v1000(ctx); err != nil {
		return fmt.Errorf("pre v10.0.0 upgrade failed: %w", err)
	}

	return nil
}
