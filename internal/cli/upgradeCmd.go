// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/upgrade"
)

// UpgradeCmd: `go-hass-agent upgrade`.
type UpgradeCmd struct{}

func (r *UpgradeCmd) Help() string {
	return showHelpTxt("upgrade-help")
}

//nolint:sloglint
func (r *UpgradeCmd) Run(opts *CmdOpts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	// Load up the contenxt.
	ctx = preferences.PathToCtx(ctx, opts.Path)
	ctx = logging.ToContext(ctx, opts.Logger)

	if err := upgrade.Run(ctx); err != nil {
		logging.FromContext(ctx).Warn(showHelpTxt("upgrade-failed-help"),
			slog.Any("error", err))

		return fmt.Errorf("upgrade failed: %w", err)
	}

	logging.FromContext(ctx).Info("All upgrades completed!")

	return nil
}
