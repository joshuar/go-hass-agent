// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent"
	"github.com/joshuar/go-hass-agent/scheduler"
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

	// ctx = preferences.PathToCtx(ctx, opts.Path)

	// // Load the preferences from file. Ignore the case where there are no
	// // existing preferences.
	// if err := preferences.Init(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
	// 	return errors.Join(ErrRunCmdFailed, err)
	// }

	err := scheduler.Start(ctx)
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	agent, err := agent.New()
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	agent.Run(ctx)

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
