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

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrInvalidRegistration = errors.New("invalid")

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func (agent *Agent) saveRegistration(resp *hass.RegistrationDetails, dev hass.DeviceInfo) error {
	// Generate an API URL from the registration info and the registration response.
	apiURL, err := generateAPIURL(agent.registrationInfo.Server, agent.registrationInfo.IgnoreOutputURLs, resp)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}

	// Generate a websocket URL from the registration info.
	websocketURL, err := generateWebsocketURL(agent.registrationInfo.Server)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}

	// Set all the preferences.
	agent.prefs.Host = agent.registrationInfo.Server
	agent.prefs.Token = agent.registrationInfo.Token
	agent.prefs.CloudhookURL = resp.CloudhookURL
	agent.prefs.RemoteUIURL = resp.RemoteUIURL
	agent.prefs.WebhookID = resp.WebhookID
	agent.prefs.Secret = resp.Secret
	agent.prefs.RestAPIURL = apiURL
	agent.prefs.WebsocketURL = websocketURL
	agent.prefs.DeviceName = dev.DeviceName()
	agent.prefs.DeviceID = dev.DeviceID()
	agent.prefs.Version = preferences.AppVersion
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
	log.Info().Msg("Registration required. Starting registration process.")

	// Display a window asking for registration details for non-headless usage.
	if !agent.headless {
		userInputDone := make(chan struct{})
		agent.ui.DisplayRegistrationWindow(ctx, agent.registrationInfo, userInputDone)
		<-userInputDone
	}

	// Validate provided registration details.
	if err := agent.registrationInfo.Validate(); err != nil {
		// if !validRegistrationSetting("server", input.Server) || !validRegistrationSetting("token", token) {
		return fmt.Errorf("failed: %w", err)
	}

	deviceInfo := device.New(preferences.AppName, preferences.AppVersion)

	// Register with Home Assistant.
	resp, err := hass.RegisterWithHass(ctx, agent.registrationInfo, deviceInfo)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	// Write registration details to config.
	if err := agent.saveRegistration(resp, deviceInfo); err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	log.Info().Msg("Successfully registered agent.")

	return nil
}

func (agent *Agent) checkRegistration(ctx context.Context, trk SensorTracker) error {
	if agent.prefs.Registered && !agent.forceRegister {
		log.Debug().Msg("Agent already registered.")

		return nil
	}

	// Agent is not registered or forced registration requested.
	if err := agent.performRegistration(ctx); err != nil {
		return err
	}

	if agent.forceRegister {
		trk.Reset()

		if err := registry.Reset(); err != nil {
			log.Warn().Err(err).Msg("Problem resetting registry.")
		}
	}

	return nil
}

func generateAPIURL(host string, ignoreURLs bool, resp *hass.RegistrationDetails) (string, error) {
	switch {
	case resp.CloudhookURL != "" && !ignoreURLs:
		return resp.CloudhookURL, nil
	case resp.RemoteUIURL != "" && resp.WebhookID != "" && !ignoreURLs:
		return resp.RemoteUIURL + hass.WebHookPath + resp.WebhookID, nil
	default:
		apiURL, err := url.Parse(host)
		if err != nil {
			return "", fmt.Errorf("unable to generate API URL: %w", err)
		}

		apiURL = apiURL.JoinPath(hass.WebHookPath, resp.WebhookID)

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
