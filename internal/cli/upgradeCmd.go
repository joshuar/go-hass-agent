// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/upgrade"
)

type UpgradeCmd struct{}

func (r *UpgradeCmd) Help() string {
	return showHelpTxt("upgrade-help")
}

func (r *UpgradeCmd) Run(ctx *Context) error {
	var logFile string

	if ctx.NoLogFile {
		logFile = ""
	} else {
		logFile = filepath.Join(xdg.ConfigHome, ctx.AppID, "upgrade.log")
	}

	logger := logging.New(ctx.LogLevel, logFile)

	upgradeCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	upgradeCtx = logging.ToContext(upgradeCtx, logger)

	if err := upgrade.Run(upgradeCtx); err != nil {
		slog.Warn(showHelpTxt("upgrade-failed-help"), slog.Any("error", err))

		return fmt.Errorf("upgrade failed: %w", err)
	}

	slog.Info("Upgrade successful!")

	return nil
}
