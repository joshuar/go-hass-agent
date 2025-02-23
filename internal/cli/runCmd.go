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
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/registry"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
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

	// Load up the contenxt.
	ctx = preferences.PathToCtx(ctx, opts.Path)
	ctx = logging.ToContext(ctx, opts.Logger)

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err := preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrRunCmdFailed, err)
	}

	err := scheduler.Start(ctx)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	// Load the registry.
	reg, err := registry.Load(opts.Path)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	client, err := hass.NewClient(ctx, reg)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	api := &API{
		hass: client,
	}

	// Run the agent.
	if err := agent.Run(ctx, opts.Headless, api); err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	return nil
}
