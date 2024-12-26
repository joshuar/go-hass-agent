// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/agent"
)

// RunCmd: `go-hass-agent run`.
type RunCmd struct{}

func (r *RunCmd) Help() string {
	return showHelpTxt("run-help")
}

func (r *RunCmd) Run(opts *CmdOpts) error {
	if err := agent.Run(
		agent.SetHeadless(opts.Headless),
		agent.SetLogger(opts.Logger),
	); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
