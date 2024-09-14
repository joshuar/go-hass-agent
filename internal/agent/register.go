// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

func checkRegistration(ctx context.Context, agentUI ui, prefs agentPreferences) error {
	if prefs.AgentRegistered() && !ForceRegister(ctx) {
		return nil
	}

	// Set the registration options as passed in from command-line.
	registrationOptions := &preferences.Registration{
		Server:         Server(ctx),
		Token:          Token(ctx),
		IgnoreHassURLs: IgnoreURLs(ctx),
	}

	// If not headless, present a UI for the user to configure options.
	if !Headless(ctx) {
		userInputDoneCh := agentUI.DisplayRegistrationWindow(ctx, registrationOptions)
		if canceled := <-userInputDoneCh; canceled {
			return ErrUserCancelledRegistration
		}
	}

	// Perform registration with given values.
	registrationDetails, err := hass.RegisterDevice(ctx, prefs.GetDeviceInfo(), registrationOptions)
	if err != nil {
		return fmt.Errorf("device registration failed: %w", err)
	}

	// Save the returned preferences.
	if err := prefs.SaveHassPreferences(registrationDetails, registrationOptions); err != nil {
		return fmt.Errorf("saving registration failed: %w", err)
	}

	// If the registration was forced, reset the sensor registry.
	if ForceRegister(ctx) {
		if err := registry.Reset(ctx); err != nil {
			logging.FromContext(ctx).Warn("Problem resetting registry.", slog.Any("error", err))
		}
	}

	logging.FromContext(ctx).Info("Agent registered.")

	return nil
}
