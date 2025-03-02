// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
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
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/registry"
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

	ctx = preferences.PathToCtx(ctx, opts.Path)

	// Load the preferences so we know what we need to reset.
	if err := preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrResetCommandFailed, err)
	}

	ctx = logging.ToContext(ctx, opts.Logger)

	// Reset agent.
	if err := agent.Reset(ctx); err != nil {
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
		slog.Warn("Reset completed with errors", slog.Any("errors", errs))
	} else {
		slog.Info("Reset completed.")
	}

	return nil
}
