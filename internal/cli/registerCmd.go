// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cli

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/agent"
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

func (r *RegisterCmd) Run(opts *CmdOpts) error {
	agentCtx, cancelFunc := newContext(opts)
	defer cancelFunc()

	gohassagent, err := agent.NewAgent(agentCtx,
		agent.Headless(opts.Headless),
		agent.WithRegistrationInfo(r.Server, r.Token, r.IgnoreURLs),
		agent.ForceRegister(r.Force))
	if err != nil {
		return fmt.Errorf("failed to run register command: %w", err)
	}

	gohassagent.Register(agentCtx)

	return nil
}
