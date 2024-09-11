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
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type RegisterCmd struct {
	Server     string `help:"Home Assistant server."`
	Token      string `help:"Personal Access Token."`
	Force      bool   `help:"Force registration."`
	IgnoreURLs bool   `help:"Ignore URLs returned by Home Assistant and use provided server for access."`
}

func (r *RegisterCmd) Help() string {
	return showHelpTxt("register-help")
}

func (r *RegisterCmd) Run(ctx *Context) error {
	agentCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	agentCtx = logging.ToContext(agentCtx, ctx.Logger)

	gohassagent, err := agent.NewAgent(agentCtx, ctx.AppID,
		agent.Headless(ctx.Headless),
		agent.WithRegistrationInfo(r.Server, r.Token, r.IgnoreURLs),
		agent.ForceRegister(r.Force))
	if err != nil {
		return fmt.Errorf("failed to run register command: %w", err)
	}

	var trk *sensor.Tracker

	if trk, err = sensor.NewTracker(); err != nil {
		return fmt.Errorf("could not start sensor tracker: %w", err)
	}

	gohassagent.Register(agentCtx, trk)

	return nil
}
