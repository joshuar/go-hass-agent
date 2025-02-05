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
	"github.com/joshuar/go-hass-agent/internal/components/tracker"
	"github.com/joshuar/go-hass-agent/internal/hass"
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

	// Load the registry.
	reg, err := registry.Load(opts.Path)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	// Load the tracker.
	trk := tracker.NewTracker()

	// Create a new hass data handler.
	dataCh, err := hass.NewDataHandler(ctx, reg, trk)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	// Run the agent.
	if err := agent.Run(ctx, opts.Headless, dataCh, trk); err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	return nil
}
