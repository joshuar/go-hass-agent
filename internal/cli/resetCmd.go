// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

type ResetCmd struct{}

func (r *ResetCmd) Help() string {
	return showHelpTxt("reset-help")
}

func (r *ResetCmd) Run(opts *CmdOpts) error {
	agentCtx, cancelFunc := newContext(opts)
	defer cancelFunc()

	var errs error

	gohassagent, err := agent.NewAgent(agentCtx,
		agent.Headless(opts.Headless))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to run reset command: %w", err))
	}

	// Reset agent.
	if err := gohassagent.Reset(agentCtx); err != nil {
		errs = errors.Join(fmt.Errorf("agent reset failed: %w", err))
	}
	// Reset registry.
	if err := registry.Reset(agentCtx); err != nil {
		errs = errors.Join(fmt.Errorf("registry reset failed: %w", err))
	}
	// Reset preferences.
	if err := preferences.Reset(agentCtx); err != nil {
		errs = errors.Join(fmt.Errorf("preferences reset failed: %w", err))
	}
	// Reset the log.
	if err := logging.Reset(agentCtx); err != nil {
		errs = errors.Join(fmt.Errorf("logging reset failed: %w", err))
	}

	if errs != nil {
		slog.Warn("Reset completed with errors", slog.Any("errors", errs))
	} else {
		slog.Info("Reset completed.")
	}

	return nil
}
