// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/upgrade"
)

// Upgrade: `go-hass-agent upgrade`.
type Upgrade struct{}

func (r *Upgrade) Help() string {
	return showHelpTxt("upgrade-help")
}

func (r *Upgrade) Run(opts *Opts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()
	ctx = slogctx.NewCtx(ctx, slog.Default())
	ctx = preferences.PathToCtx(ctx, opts.Path)

	if err := upgrade.Run(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn(showHelpTxt("upgrade-failed-help"),
			slog.Any("error", err))

		return fmt.Errorf("upgrade failed: %w", err)
	}

	slogctx.FromCtx(ctx).Info("All upgrades completed!")

	return nil
}
