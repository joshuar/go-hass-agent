// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrInvalidRegistration = errors.New("invalid")

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func (agent *Agent) saveRegistration(hassPrefs *preferences.Hass) error {
	var err error

	// Copy over existing preferences.
	hassPrefs.IgnoreHassURLs = agent.prefs.Hass.IgnoreHassURLs
	// Copy new hass preferences to agent preferences
	agent.prefs.Hass = hassPrefs
	// Add the generated URLS
	// Generate an API URL.
	agent.prefs.Hass.RestAPIURL, err = generateAPIURL(agent.prefs.Registration.Server, hassPrefs)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}
	// Generate a websocket URL.
	agent.prefs.Hass.WebsocketURL, err = generateWebsocketURL(agent.prefs.Registration.Server)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}
	// Set agent as registered
	agent.prefs.Registered = true
	// Save the preferences to disk.
	err = agent.prefs.Save()
	if err != nil {
		return fmt.Errorf("unable to save preferences: %w", err)
	}

	return nil
}

// performRegistration runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config.
func (agent *Agent) performRegistration(ctx context.Context) error {
	agent.logger.Info("Registration required. Starting registration process.")

	// Display a window asking for registration details for non-headless usage.
	if !agent.headless {
		userInputDone := make(chan struct{})
		agent.ui.DisplayRegistrationWindow(ctx, agent.prefs, userInputDone)
		<-userInputDone
	}

	// Validate provided registration details.
	if err := agent.prefs.Registration.Validate(); err != nil {
		// if !validRegistrationSetting("server", input.Server) || !validRegistrationSetting("token", token) {
		return fmt.Errorf("failed: %w", err)
	}

	// Register with Home Assistant.
	resp, err := hass.RegisterWithHass(ctx, agent.prefs.Device, agent.prefs.Registration)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	// Write registration details to config.
	if err := agent.saveRegistration(resp); err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	agent.logger.Info("Successfully registered agent.")

	return nil
}

func (agent *Agent) checkRegistration(ctx context.Context, trk SensorTracker) error {
	if agent.prefs.Registered && !agent.forceRegister {
		agent.logger.Debug("Agent is already registered. Skipping.")

		return nil
	}

	// Agent is not registered or forced registration requested.
	if err := agent.performRegistration(ctx); err != nil {
		return err
	}

	if agent.forceRegister {
		trk.Reset()

		if err := registry.Reset(); err != nil {
			agent.logger.Warn("Problem resetting registry.", "error", err.Error())
		}
	}

	return nil
}

func generateAPIURL(server string, prefs *preferences.Hass) (string, error) {
	switch {
	case prefs.CloudhookURL != "" && !prefs.IgnoreHassURLs:
		return prefs.CloudhookURL, nil
	case prefs.RemoteUIURL != "" && prefs.WebhookID != "" && !prefs.IgnoreHassURLs:
		return prefs.RemoteUIURL + hass.WebHookPath + prefs.WebhookID, nil
	default:
		apiURL, err := url.Parse(server)
		if err != nil {
			return "", fmt.Errorf("unable to generate API URL: %w", err)
		}

		apiURL = apiURL.JoinPath(hass.WebHookPath, prefs.WebhookID)

		return apiURL.String(), nil
	}
}

func generateWebsocketURL(host string) (string, error) {
	websocketURL, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("unable to generate websocket URL: %w", err)
	}

	switch websocketURL.Scheme {
	case "https":
		websocketURL.Scheme = "wss"
	case "http":
		websocketURL.Scheme = "ws"
	case "wss":
	default:
		websocketURL.Scheme = "ws"
	}

	websocketURL = websocketURL.JoinPath(hass.WebsocketPath)

	return websocketURL.String(), nil
}
