// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// ResetCmd: `go-hass-agent reset`.
type ResetCmd struct{}

func (r *ResetCmd) Help() string {
	return showHelpTxt("reset-help")
}

func (r *ResetCmd) Run(opts *CmdOpts) error {
	var errs error

	// Reset agent.
	if err := agent.Reset(opts.Logger,
		agent.SetHeadless(opts.Headless),
	); err != nil {
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
