// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run github.com/matryer/moq -out register_mocks_test.go . registrationPrefs
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// checkRegistration retrieves any registration information passed on the
// command-line and then checks to see if the agent needs to register to Home
// Assistant. If it does, it will perform the registration via either a
// graphical (user-prompted) or non-graphical (automatic) process.
func checkRegistration(ctx context.Context, agentUI ui) error {
	// Retrieve request options passed on command-line from context.
	request := RegistrationFromCtx(ctx)
	if request == nil {
		request = &preferences.Registration{}
	}

	if preferences.Registered() && !request.ForceRegister {
		logging.FromContext(ctx).Debug("Already registered and forced registration not requested.")
		return nil
	}

	// If not headless, present a UI for the user to configure options.
	if !HeadlessFromCtx(ctx) {
		userInputDoneCh := agentUI.DisplayRegistrationWindow(ctx, request)
		if canceled := <-userInputDoneCh; canceled {
			return errors.New("user canceled registration")
		}
	}

	// Perform registration with given values.
	result, err := hass.RegisterDevice(ctx, request)
	if err != nil {
		return fmt.Errorf("device registration failed: %w", err)
	}
	// Save the returned preferences.
	if err := preferences.SetHassPreferences(result, request); err != nil {
		return fmt.Errorf("saving registration failed: %w", err)
	}
	// Set registration status.
	if err := preferences.SetRegistered(true); err != nil {
		return fmt.Errorf("saving registration failed: %w", err)
	}
	// Save preferences to disk.
	if err := preferences.Save(); err != nil {
		return fmt.Errorf("saving registration failed: %w", err)
	}

	// If the registration was forced, reset the sensor registry.
	if request.ForceRegister {
		if err := registry.Reset(); err != nil {
			logging.FromContext(ctx).Warn("Problem resetting registry.",
				slog.Any("error", err))
		}
	}

	logging.FromContext(ctx).Info("Agent registered.")

	return nil
}
