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
	"github.com/joshuar/go-hass-agent/device"
	"github.com/joshuar/go-hass-agent/hass"
)

// Register represents the options for the `register` command.
type Register struct {
	hass.RegistrationRequest

	Force bool `help:"Force registration."`
}

// Run processes the register command.
func (r *Register) Run(_ *Opts) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()
	ctx = slogctx.NewCtx(ctx, slog.Default())

	// Create an agent instance.
	agent, err := agent.New()
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}

	// Don't continue if agent is registered unless force option is set.
	if agent.IsRegistered() && !r.Force {
		slogctx.FromCtx(ctx).Warn("Already registered and force not set.")
		return nil
	}

	// Validate registration options.
	valid, err := r.Valid()
	if !valid || err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}

	// Get the device config.
	deviceCfg, err := device.GetConfig()
	if err != nil {
		return fmt.Errorf("unable to register: get device details failed: %w", err)
	}

	// Perform registration.
	err = hass.Register(ctx, deviceCfg.ID, &r.RegistrationRequest)
	if err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}

	// If force option set, reset the agent.
	if r.Force {
		err := hass.Reset()
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not reset registry state.",
				slog.Any("error", err))
		}
		agent.Reset(ctx)
	}

	// Register the agent.
	agent.Register(ctx)

	slogctx.FromCtx(ctx).Info("Agent registered!")

	return nil
}
