// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/app"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/registry"
)

var ErrResetCommandFailed = errors.New("reset command failed")

// Reset: `go-hass-agent reset`.
type Reset struct{}

func (r *Reset) Help() string {
	return showHelpTxt("reset-help")
}

func (r *Reset) Run(opts *Opts) error {
	var errs error

	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()
	ctx = slogctx.NewCtx(ctx, slog.Default())
	ctx = preferences.PathToCtx(ctx, opts.Path)

	// Load the preferences so we know what we need to reset.
	if err := preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrResetCommandFailed, err)
	}

	// Reset agent.
	if err := app.Reset(ctx); err != nil {
		errs = errors.Join(fmt.Errorf("agent reset failed: %w", err))
	}
	// Reset registry.
	if err := registry.Reset(opts.Path); err != nil {
		errs = errors.Join(fmt.Errorf("registry reset failed: %w", err))
	}
	// Reset preferences.
	if err := preferences.Reset(ctx); err != nil {
		errs = errors.Join(fmt.Errorf("preferences reset failed: %w", err))
	}
	// Reset the log.
	if err := logging.Reset(opts.Path); err != nil {
		errs = errors.Join(fmt.Errorf("logging reset failed: %w", err))
	}

	if errs != nil {
		slogctx.FromCtx(ctx).Warn("Reset completed with errors", slog.Any("errors", errs))
	} else {
		slogctx.FromCtx(ctx).Info("Reset completed.")
	}

	return nil
}
