// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"net/url"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func saveRegistration(input *hass.RegistrationInput, resp *hass.RegistrationDetails, dev hass.DeviceInfo) error {
	return preferences.Save(
		preferences.SetHost(input.Server),
		preferences.SetToken(input.Token),
		preferences.SetCloudhookURL(resp.CloudhookURL),
		preferences.SetRemoteUIURL(resp.RemoteUIURL),
		preferences.SetWebhookID(resp.WebhookID),
		preferences.SetSecret(resp.Secret),
		preferences.SetRestAPIURL(generateAPIURL(input.Server, input.IgnoreOutputURLs, resp)),
		preferences.SetWebsocketURL(generateWebsocketURL(input.Server)),
		preferences.SetDeviceName(dev.DeviceName()),
		preferences.SetDeviceID(dev.DeviceID()),
		preferences.SetVersion(preferences.AppVersion),
		preferences.SetRegistered(true),
	)
}

// performRegistration runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config.
func (agent *Agent) performRegistration(ctx context.Context, server, token string) error {
	log.Info().Msg("Registration required. Starting registration process.")

	input := &hass.RegistrationInput{
		Server: server,
		Token:  token,
	}

	// Display a window asking for registration details for non-headless usage.
	if !agent.Options.Headless {
		userInputDone := make(chan struct{})
		agent.ui.DisplayRegistrationWindow(ctx, input, userInputDone)
		<-userInputDone
	}

	// Validate provided registration details.
	if input.Validate() != nil {
		// if !validRegistrationSetting("server", input.Server) || !validRegistrationSetting("token", token) {
		return errors.New("cannot register, invalid host and/or token")
	}

	// Register with Home Assistant.
	device := newDevice(ctx)
	resp, err := hass.RegisterWithHass(ctx, input, device)
	if err != nil {
		return err
	}

	// Write registration details to config.
	if err := saveRegistration(input, resp, device); err != nil {
		return errors.New("could not save registration")
	}

	log.Info().Msg("Successfully registered agent.")
	return nil
}

func (agent *Agent) checkRegistration(trk SensorTracker) error {
	prefs, err := preferences.Load()
	if err != nil && !os.IsNotExist(err) {
		return errors.New("could not load preferences")
	}
	if prefs.Registered && !agent.Options.ForceRegister {
		log.Debug().Msg("Agent already registered.")
		return nil
	}

	// Agent is not registered or forced registration requested.
	if err := agent.performRegistration(context.Background(), agent.Options.Server, agent.Options.Token); err != nil {
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

func generateAPIURL(host string, ignoreURLs bool, resp *hass.RegistrationDetails) string {
	switch {
	case resp.CloudhookURL != "" && !ignoreURLs:
		return resp.CloudhookURL
	case resp.RemoteUIURL != "" && resp.WebhookID != "" && !ignoreURLs:
		return resp.RemoteUIURL + hass.WebHookPath + resp.WebhookID
	default:
		u, _ := url.Parse(host)
		u = u.JoinPath(hass.WebHookPath, resp.WebhookID)
		return u.String()
	}
}

func generateWebsocketURL(host string) string {
	// TODO: look into websocket http upgrade method
	u, err := url.Parse(host)
	if err != nil {
		log.Warn().Err(err).Msg("Could not parse URL.")
		return ""
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	case "ws":
		// nothing to do
	case "wss":
		// nothing to do
	default:
		u.Scheme = "ws"
	}
	u = u.JoinPath(hass.WebsocketPath)
	return u.String()
}
