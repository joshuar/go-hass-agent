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

var ErrUserCancelledRegistration = errors.New("user canceled registration")

func checkRegistration(ctx context.Context, agentUI ui) error {
	if preferences.Registered() && !options.forceRegister {
		return nil
	}

	// Set the registration options as passed in from command-line.
	registrationOptions := &preferences.Registration{
		Server:         options.registrationServer,
		Token:          options.registrationToken,
		IgnoreHassURLs: options.ignoreHassURLs,
	}

	// If not headless, present a UI for the user to configure options.
	if !options.headless {
		userInputDoneCh := agentUI.DisplayRegistrationWindow(ctx, registrationOptions)
		if canceled := <-userInputDoneCh; canceled {
			return ErrUserCancelledRegistration
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
	if options.forceRegister {
		if err := registry.Reset(); err != nil {
			logging.FromContext(ctx).Warn("Problem resetting registry.", slog.Any("error", err))
		}
	}

	logging.FromContext(ctx).Info("Agent registered.")

	return nil
}
