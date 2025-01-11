// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cli

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

var ErrRunCmdFailed = errors.New("run command failed")

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

	dataCh, err := hass.NewDataHandler(ctx)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	if err := agent.Run(ctx, dataCh); err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	return nil
}
