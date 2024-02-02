// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"net/url"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/preferences"

	"github.com/go-playground/validator/v10"
)

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func saveRegistration(server, token string, resp *api.RegistrationResponse, dev api.DeviceInfo) error {
	return preferences.Save(
		preferences.Host(server),
		preferences.Token(token),
		preferences.CloudhookURL(resp.CloudhookURL),
		preferences.RemoteUIURL(resp.RemoteUIURL),
		preferences.WebhookID(resp.WebhookID),
		preferences.Secret(resp.Secret),
		preferences.RestAPIURL(generateAPIURL(server, resp)),
		preferences.WebsocketURL(generateWebsocketURL(server)),
		preferences.Name(dev.DeviceName()),
		preferences.ID(dev.DeviceID()),
		preferences.Version(preferences.AppVersion),
		preferences.Registered(true),
	)
}

// performRegistration runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config.
func (agent *Agent) performRegistration(ctx context.Context, server, token string) {
	log.Info().Msg("Registration required. Starting registration process.")

	// Display a window asking for registration details for non-headless usage.
	if !agent.Options.Headless {
		userInputDone := make(chan struct{})
		agent.ui.DisplayRegistrationWindow(ctx, &server, &token, userInputDone)
		<-userInputDone
	}

	// Validate provided registration details.
	if !validRegistrationSetting("server", server) || !validRegistrationSetting("token", token) {
		log.Fatal().Msg("Cannot register, invalid host and/or token.")
	}

	// Register with Home Assistant.
	device := newDevice(ctx)
	resp, err := api.RegisterWithHass(ctx, server, token, device)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not register with Home Assistant.")
	}

	// Write registration details to config.
	if err := saveRegistration(server, token, resp, device); err != nil {
		log.Fatal().Err(err).Msg("Could not save registration.")
	}

	log.Info().Msg("Successfully registered agent.")
}

func (agent *Agent) checkRegistration(trk SensorTracker, prefs *preferences.Preferences) {
	// If the agent is not registered (or force registration requested) run a
	// registration flow
	if !prefs.Registered || agent.Options.ForceRegister {
		agent.performRegistration(context.Background(), agent.Options.Server, agent.Options.Token)
		if agent.Options.ForceRegister {
			trk.Reset()
		}
	} else {
		log.Debug().Msg("Agent already registered.")
	}
}

func validRegistrationSetting(key, value string) bool {
	if value == "" {
		return false
	}
	validate := validator.New()
	check := func(value string, validation string) bool {
		if err := validate.Var(value, validation); err != nil {
			return false
		}
		return true
	}
	switch key {
	case "server":
		return check(value, "required,http_url")
	case "token":
		return check(value, "required")
	default:
		log.Warn().Msgf("Unexpected key %s with value %s", key, value)
		return false
	}
}

func generateAPIURL(host string, resp *api.RegistrationResponse) string {
	switch {
	case resp.CloudhookURL != "":
		return resp.CloudhookURL
	case resp.RemoteUIURL != "" && resp.WebhookID != "":
		return resp.RemoteUIURL + api.WebHookPath + resp.WebhookID
	case resp.WebhookID != "":
		u, _ := url.Parse(host)
		u = u.JoinPath(api.WebHookPath, resp.WebhookID)
		return u.String()
	default:
		return ""
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
	u = u.JoinPath(api.WebsocketPath)
	return u.String()
}
