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

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type ResetCmd struct{}

func (r *ResetCmd) Help() string {
	return showHelpTxt("reset-help")
}

func (r *ResetCmd) Run(ctx *Context) error {
	agentCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	agentCtx = logging.ToContext(agentCtx, ctx.Logger)

	var errs error

	gohassagent, err := agent.NewAgent(agentCtx, ctx.AppID,
		agent.Headless(ctx.Headless))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to run reset command: %w", err))
	}

	// Reset agent.
	if err := gohassagent.Reset(agentCtx); err != nil {
		errs = errors.Join(fmt.Errorf("agent reset failed: %w", err))
	}
	// Reset registry.
	if err := registry.Reset(gohassagent.GetRegistryPath()); err != nil {
		errs = errors.Join(fmt.Errorf("registry reset failed: %w", err))
	}
	// Reset preferences.
	if err := preferences.Reset(gohassagent.GetPreferencesPath()); err != nil {
		errs = errors.Join(fmt.Errorf("preferences reset failed: %w", err))
	}
	// Reset the log.
	// if err := logging.Reset(logFile); err != nil {
	// 	errs = errors.Join(fmt.Errorf("logging reset failed: %w", err))
	// }

	if errs != nil {
		slog.Warn("Reset completed with errors", slog.Any("errors", errs))
	} else {
		slog.Info("Reset completed.")
	}

	return nil
}
