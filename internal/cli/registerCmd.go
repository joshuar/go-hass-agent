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
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

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

	ctx = logging.ToContext(ctx, opts.Logger)
	ctx = agent.HeadlessToCtx(ctx, opts.Headless)
	ctx = agent.RegistrationToCtx(ctx, preferences.Registration{
		Server:         r.Server,
		Token:          r.Token,
		IgnoreHassURLs: r.IgnoreURLs,
		ForceRegister:  r.Force,
	})

	if err := agent.Register(ctx); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
