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

	agentCtx = agent.LoadCtx(agentCtx,
		agent.SetHeadless(opts.Headless),
		agent.SetRegistrationInfo(r.Server, r.Token, r.IgnoreURLs),
		agent.SetForceRegister(r.Force))

	if err := agent.Register(agentCtx); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
