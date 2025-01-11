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
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err := preferences.Load(); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrRunCmdFailed, err)
	}

	// Load up the context for the agent.
	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = agent.HeadlessToCtx(ctx, opts.Headless)

	// Create a new hass data handler.
	dataCh, err := hass.NewDataHandler(ctx)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	// Run the agent.
	if err := agent.Run(ctx, dataCh); err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	return nil
}
