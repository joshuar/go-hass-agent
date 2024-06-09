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
	"os"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrInvalidRegistration = errors.New("invalid")

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func saveRegistration(input *hass.RegistrationInput, resp *hass.RegistrationDetails, dev hass.DeviceInfo) error {
	apiURL, err := generateAPIURL(input.Server, input.IgnoreOutputURLs, resp)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}

	websocketURL, err := generateWebsocketURL(input.Server)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}

	err = preferences.Save(
		preferences.SetHost(input.Server),
		preferences.SetToken(input.Token),
		preferences.SetCloudhookURL(resp.CloudhookURL),
		preferences.SetRemoteUIURL(resp.RemoteUIURL),
		preferences.SetWebhookID(resp.WebhookID),
		preferences.SetSecret(resp.Secret),
		preferences.SetRestAPIURL(apiURL),
		preferences.SetWebsocketURL(websocketURL),
		preferences.SetDeviceName(dev.DeviceName()),
		preferences.SetDeviceID(dev.DeviceID()),
		preferences.SetVersion(preferences.AppVersion),
		preferences.SetRegistered(true),
	)
	if err != nil {
		return fmt.Errorf("unable to save registration: %w", err)
	}

	return nil
}

// performRegistration runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config.
func (agent *Agent) performRegistration(ctx context.Context) error {
	log.Info().Msg("Registration required. Starting registration process.")

	input := &hass.RegistrationInput{
		Server:           agent.Options.Server,
		Token:            agent.Options.Token,
		IgnoreOutputURLs: agent.Options.IgnoreURLs,
	}

	// Display a window asking for registration details for non-headless usage.
	if !agent.Options.Headless {
		userInputDone := make(chan struct{})
		agent.ui.DisplayRegistrationWindow(ctx, input, userInputDone)
		<-userInputDone
	}

	// Validate provided registration details.
	if err := input.Validate(); err != nil {
		// if !validRegistrationSetting("server", input.Server) || !validRegistrationSetting("token", token) {
		return fmt.Errorf("failed: %w", err)
	}

	device := newDevice(ctx)

	// Register with Home Assistant.
	resp, err := hass.RegisterWithHass(ctx, input, device)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	// Write registration details to config.
	if err := saveRegistration(input, resp, device); err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	log.Info().Msg("Successfully registered agent.")

	return nil
}

func (agent *Agent) checkRegistration(trk SensorTracker) error {
	prefs, err := preferences.Load()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not load preferences: %w", err)
	}

	if prefs.Registered && !agent.Options.ForceRegister {
		log.Debug().Msg("Agent already registered.")

		return nil
	}

	// Agent is not registered or forced registration requested.
	if err := agent.performRegistration(context.Background()); err != nil {
		return err
	}

	if agent.Options.ForceRegister {
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
