// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/scheduler"
	"github.com/joshuar/go-hass-agent/server"
)

// Run is the command-line option for running the agent.
type Run struct {
	ServerHTTPSCert string `help:"Path to cert file for using https for web server component."`
	ServerHTTPSKey  string `help:"Path to key file for using https for web server component."`
	ServerHostname  string `help:"Hostname that web server component will listen on."          default:"localhost"`
	ServerPort      string `help:"Port that web server component will listen on."              default:"8223"`
}

// Help shows a help message about the run command.
func (r *Run) Help() string {
	return "Run Go Hass Agent with the given options."
}

// Run starts the agent.
func (r *Run) Run(opts *Opts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()
	ctx = slogctx.NewCtx(ctx, slog.Default())

	err := config.Init()
	if err != nil && !errors.Is(err, config.ErrLoadConfig) {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Start scheduler.
	err = scheduler.Start(ctx)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Configure agent.
	agent, err := agent.New()
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Configure web server.
	server, err := server.New(ctx,
		opts.StaticContent,
		agent,
		server.WithHost(r.ServerHostname),
		server.WithPort(r.ServerPort),
		server.WithCertFile(r.ServerHTTPSCert),
		server.WithKeyFile(r.ServerHTTPSKey),
	)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Start web server.
	go func() {
		if err = server.Start(ctx); err != nil {
			panic(fmt.Errorf("unable to run: %w", err))
		}
	}()

	if !agent.IsRegistered() {
		xdgOpen, err := exec.LookPath("xdg-open")
		if err != nil {
			slogctx.FromCtx(ctx).
				Info("Agent is not registered. Please open your web browser to " + server.ShowAddress() + "/register to register the agent with Home Assistant")
		} else {
			_, err := exec.CommandContext(ctx, xdgOpen, server.ShowAddress()).Output()
			if err != nil {
				slogctx.FromCtx(ctx).
					Info("Agent is not registered. Please open your web browser to " + server.ShowAddress() + "/register to register the agent with Home Assistant")
			}
		}
	}

	// Start agent.
	err = agent.Run(ctx)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	return nil
}
