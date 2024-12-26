// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/upgrade"
)

// UpgradeCmd: `go-hass-agent upgrade`.
type UpgradeCmd struct{}

func (r *UpgradeCmd) Help() string {
	return showHelpTxt("upgrade-help")
}

func (r *UpgradeCmd) Run(_ *CmdOpts) error {
	if err := upgrade.Run(); err != nil {
		if errors.Is(err, upgrade.ErrNoPrevConfig) {
			slog.Info("No previous installation found. Nothing to do!")
			return nil
		}

		slog.Warn(showHelpTxt("upgrade-failed-help"), slog.Any("error", err)) //nolint:sloglint

		return fmt.Errorf("upgrade failed: %w", err)
	}

	slog.Info("Upgrade successful!")

	return nil
}
