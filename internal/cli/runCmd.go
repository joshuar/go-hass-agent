// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type RunCmd struct{}

func (r *RunCmd) Help() string {
	return showHelpTxt("run-help")
}

func (r *RunCmd) Run(ctx *Context) error {
	agentCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	agentCtx = logging.ToContext(agentCtx, ctx.Logger)

	gohassagent, err := agent.NewAgent(agentCtx, ctx.AppID,
		agent.Headless(ctx.Headless))
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	var trk *sensor.Tracker

	reg, err := registry.Load(gohassagent.GetRegistryPath())
	if err != nil {
		return fmt.Errorf("could not start registry: %w", err)
	}

	if trk, err = sensor.NewTracker(); err != nil {
		return fmt.Errorf("could not start sensor tracker: %w", err)
	}

	if err := gohassagent.Run(agentCtx, trk, reg); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
