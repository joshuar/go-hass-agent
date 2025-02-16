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

var ErrRegisterCmdFailed = errors.New("register command failed")

// RegisterCmd: `go-hass-agent register`.
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
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	ctx = preferences.PathToCtx(ctx, opts.Path)

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err := preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	// Load up the context for the agent.
	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = preferences.RegistrationToCtx(ctx, preferences.Registration{
		Server:         r.Server,
		Token:          r.Token,
		IgnoreHassURLs: r.IgnoreURLs,
		ForceRegister:  r.Force,
	})

	// Load the registry.
	reg, err := registry.Load(opts.Path)
	if err != nil {
		return errors.Join(ErrRunCmdFailed, err)
	}

	// Load the tracker.
	trk := tracker.NewTracker()

	api := &API{
		hass: hass.NewClient(ctx, trk, reg),
	}

	// Run the agent.
	if err := agent.Register(ctx, opts.Headless, api); err != nil {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	return nil
}
