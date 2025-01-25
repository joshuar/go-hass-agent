// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package upgrade

import (
	"context"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
)

func Run(ctx context.Context) error {
	logging.FromContext(ctx).Info("Checking for and attempting pre v10.0.0 upgrades...")
	// Perform pre v10.0.0 upgrades...
	if err := v1000(ctx); err != nil {
		return fmt.Errorf("pre v10.0.0 upgrade failed: %w", err)
	}

	return nil
}
