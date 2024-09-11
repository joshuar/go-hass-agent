// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/upgrade"
)

type UpgradeCmd struct{}

func (r *UpgradeCmd) Help() string {
	return showHelpTxt("upgrade-help")
}

func (r *UpgradeCmd) Run(ctx *Context) error {
	upgradeCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	upgradeCtx = logging.ToContext(upgradeCtx, ctx.Logger)

	if err := upgrade.Run(upgradeCtx); err != nil {
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
