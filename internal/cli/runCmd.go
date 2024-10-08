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

type RunCmd struct{}

func (r *RunCmd) Help() string {
	return showHelpTxt("run-help")
}

func (r *RunCmd) Run(opts *CmdOpts) error {
	agentCtx, cancelFunc := newContext(opts)
	defer cancelFunc()

	agentCtx = agent.LoadCtx(agentCtx, agent.SetHeadless(opts.Headless))

	if err := agent.Run(agentCtx); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
