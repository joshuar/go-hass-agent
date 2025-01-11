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
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err := preferences.Load(); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	// Load up the context for the agent.
	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = agent.HeadlessToCtx(ctx, opts.Headless)
	ctx = agent.RegistrationToCtx(ctx, preferences.Registration{
		Server:         r.Server,
		Token:          r.Token,
		IgnoreHassURLs: r.IgnoreURLs,
		ForceRegister:  r.Force,
	})

	// Run the agent.
	if err := agent.Register(ctx); err != nil {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	return nil
}
