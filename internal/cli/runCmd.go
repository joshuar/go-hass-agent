// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

// RunCmd: `go-hass-agent run`.
type RunCmd struct{}

func (r *RunCmd) Help() string {
	return showHelpTxt("run-help")
}

func (r *RunCmd) Run(opts *CmdOpts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = agent.HeadlessToCtx(ctx, opts.Headless)

	if err := agent.Run(ctx); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
