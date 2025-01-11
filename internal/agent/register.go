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

func checkRegistration(ctx context.Context, agentUI ui) error {
	if RegistrationFromCtx(ctx) != nil {
		if preferences.Registered() && !RegistrationFromCtx(ctx).ForceRegister {
			return nil
		}
	}

	if preferences.Registered() {
		return nil
	}

	// Retrieve registration options passed on command-line from context.
	registrationOptions := RegistrationFromCtx(ctx)
	if registrationOptions == nil {
		registrationOptions = &preferences.Registration{}
	}

	// If not headless, present a UI for the user to configure options.
	if !HeadlessFromCtx(ctx) {
		userInputDoneCh := agentUI.DisplayRegistrationWindow(ctx, registrationOptions)
		if canceled := <-userInputDoneCh; canceled {
			return errors.New("user canceled registration")
		}
	}

	// Perform registration with given values.
	registrationDetails, err := hass.RegisterDevice(ctx, registrationOptions)
	if err != nil {
		return fmt.Errorf("device registration failed: %w", err)
	}
	// Save the returned preferences.
	if err := preferences.SetHassPreferences(registrationDetails, registrationOptions); err != nil {
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
	if registrationOptions.ForceRegister {
		if err := registry.Reset(); err != nil {
			logging.FromContext(ctx).Warn("Problem resetting registry.",
				slog.Any("error", err))
		}
	}

	logging.FromContext(ctx).Info("Agent registered.")

	return nil
}
