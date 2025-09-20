// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/scheduler"
	"github.com/joshuar/go-hass-agent/server"
)

// Run: `go-hass-agent run`.
type Run struct{}

func (r *Run) Help() string {
	return "Run Go Hass Agent with the given options."
}

func (r *Run) Run(opts *Opts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()
	ctx = slogctx.NewCtx(ctx, slog.Default())

	if err := config.Init(); err != nil && !errors.Is(err, config.ErrLoadConfig) {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Start scheduler.
	err := scheduler.Start(ctx)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Configure agent.
	agent, err := agent.New()
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Configure web server.
	server, err := server.New(opts.StaticContent, agent)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Start web server.
	err = server.Start(ctx)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Start agent.
	err = agent.Run(ctx)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// client, err := hass.NewClient(ctx, reg)
	// if err != nil {
	// 	return errors.Join(ErrRunCmdFailed, err)
	// }

	// api := &API{
	// 	hass: client,
	// }

	// // Run the agent.
	// if err := app.Run(ctx, opts.Headless, api); err != nil {
	// 	return errors.Join(ErrRunCmdFailed, err)
	// }

	return nil
}
