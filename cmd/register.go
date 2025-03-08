// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/agent"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

var ErrRegisterCmdFailed = errors.New("register command failed")

// Register: `go-hass-agent register`.
type Register struct {
	Server     string `help:"Home Assistant server."`
	Token      string `help:"Personal Access Token."`
	Force      bool   `help:"Force registration."`
	IgnoreURLs bool   `help:"Ignore URLs returned by Home Assistant and use provided server for access."`
}

func (r *Register) Help() string {
	return showHelpTxt("register-help")
}

func (r *Register) Run(opts *Opts) error {
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

	// Run the agent.
	if err := agent.Register(ctx, opts.Headless); err != nil {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	return nil
}
