// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrResetCommandFailed = errors.New("reset command failed")

// ResetCmd: `go-hass-agent reset`.
type ResetCmd struct{}

func (r *ResetCmd) Help() string {
	return showHelpTxt("reset-help")
}

func (r *ResetCmd) Run(opts *CmdOpts) error {
	var errs error

	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	// Load the preferences so we know what we need to reset.
	if err := preferences.Load(); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrResetCommandFailed, err)
	}

	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = preferences.HeadlessToCtx(ctx, opts.Headless)

	// Reset agent.
	if err := agent.Reset(ctx); err != nil {
		errs = errors.Join(fmt.Errorf("agent reset failed: %w", err))
	}
	// Reset registry.
	if err := registry.Reset(); err != nil {
		errs = errors.Join(fmt.Errorf("registry reset failed: %w", err))
	}
	// Reset preferences.
	if err := preferences.Reset(); err != nil {
		errs = errors.Join(fmt.Errorf("preferences reset failed: %w", err))
	}
	// Reset the log.
	if err := logging.Reset(); err != nil {
		errs = errors.Join(fmt.Errorf("logging reset failed: %w", err))
	}

	if errs != nil {
		slog.Warn("Reset completed with errors", slog.Any("errors", errs))
	} else {
		slog.Info("Reset completed.")
	}

	return nil
}
