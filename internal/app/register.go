// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/components/registry"
	"github.com/joshuar/go-hass-agent/internal/hass/registration"
)

var ErrRegister = errors.New("error registering agent")

// checkRegistration retrieves any registration information passed on the
// command-line and then checks to see if the agent needs to register to Home
// Assistant. If it does, it will perform the registration via either a
// graphical (user-prompted) or non-graphical (automatic) process.
func checkRegistration(ctx context.Context, app *App, headless bool) error {
	// Retrieve request options passed on command-line from context.
	request := preferences.RegistrationFromCtx(ctx)
	if request == nil {
		request = &preferences.Registration{}
	}

	if preferences.Registered() && !request.ForceRegister {
		slogctx.FromCtx(ctx).Debug("Already registered and forced registration not requested.")
		return nil
	}

	// If not headless, present a UI for the user to configure options.
	if !headless {
		userInputDoneCh := app.ui.DisplayRegistrationWindow(ctx, request)
		if canceled := <-userInputDoneCh; canceled {
			return fmt.Errorf("%w: user cancelled registration", ErrRegister)
		}
	}

	// Perform registration with given values.
	err := registration.RegisterDevice(ctx, request)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRegister, err)
	}

	// If the registration was forced, reset the sensor registry.
	if request.ForceRegister {
		if err := registry.Reset(preferences.PathFromCtx(ctx)); err != nil {
			slogctx.FromCtx(ctx).Warn("Problem resetting registry.",
				slog.Any("error", err))
		}
	}

	slogctx.FromCtx(ctx).Info("Agent registered.")

	return nil
}

func Register(ctx context.Context) error {
	if err := checkRegistration(ctx, nil, true); err != nil {
		return errors.Join(ErrRegister, err)
	}

	return nil
}
