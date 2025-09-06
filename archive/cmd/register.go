// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/app"
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
	ctx = slogctx.NewCtx(ctx, slog.Default())
	ctx = preferences.PathToCtx(ctx, opts.Path)

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err := preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	// Load up the context for the agent.
	ctx = preferences.RegistrationToCtx(ctx, preferences.Registration{
		Server:         r.Server,
		Token:          r.Token,
		IgnoreHassURLs: r.IgnoreURLs,
		ForceRegister:  r.Force,
	})

	// Run the agent.
	if err := app.Register(ctx); err != nil {
		return errors.Join(ErrRegisterCmdFailed, err)
	}

	return nil
}
