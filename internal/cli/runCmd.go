// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cli

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
)

type RunCmd struct{}

func (r *RunCmd) Help() string {
	return showHelpTxt("run-help")
}

func (r *RunCmd) Run(opts *CmdOpts) error {
	agentCtx, cancelFunc := newContext(opts)
	defer cancelFunc()

	gohassagent, err := agent.NewAgent(agentCtx, agent.Headless(opts.Headless))
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	var trk *sensor.Tracker

	reg, err := registry.Load(agentCtx)
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
