// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"net/url"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"

	"github.com/go-playground/validator/v10"
)

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func saveRegistration(cfg config.Config, server string, resp *api.RegistrationResponse, dev api.DeviceInfo) {
	checkFatal := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("Could not save registration.")
		}
	}

	if resp.CloudhookURL != "" {
		checkFatal(cfg.Set(config.PrefCloudhookURL, resp.CloudhookURL))
	}
	if resp.RemoteUIURL != "" {
		checkFatal(cfg.Set(config.PrefRemoteUIURL, resp.RemoteUIURL))
	}
	if resp.Secret != "" {
		checkFatal(cfg.Set(config.PrefSecret, resp.Secret))
	}
	if resp.WebhookID != "" {
		checkFatal(cfg.Set(config.PrefWebhookID, resp.WebhookID))
	}
	checkFatal(cfg.Set(config.PrefAPIURL, generateAPIURL(server, resp)))
	checkFatal(cfg.Set(config.PrefWebsocketURL, generateWebsocketURL(server)))
	checkFatal(cfg.Set(config.PrefDeviceName, dev.DeviceName()))
	checkFatal(cfg.Set(config.PrefDeviceID, dev.DeviceID()))
	checkFatal(cfg.Set(config.PrefRegistered, true))
	checkFatal(cfg.Set(config.PrefVersion, config.AppVersion))
}

// performRegistration runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config.
func (agent *Agent) performRegistration(ctx context.Context, server, token string, cfg config.Config) {
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
	saveRegistration(cfg, server, resp, device)
	if err = cfg.Set(config.PrefHost, server); err != nil {
		log.Fatal().Err(err).Msg("Could not set host preference.")
	}
	if err = cfg.Set(config.PrefToken, token); err != nil {
		log.Fatal().Err(err).Msg("Could not set token preference.")
	}

	// Ensure new config is valid.
	if err = config.ValidateConfig(cfg); err != nil {
		log.Fatal().Err(err).Msg("Could not validate config after registration.")
	}
	log.Info().Msg("Successfully registered agent.")
}

func (agent *Agent) checkRegistration(t *tracker.SensorTracker, c config.Config) {
	var registered bool
	if err := c.Get(config.PrefRegistered, &registered); err != nil {
		log.Fatal().Err(err).Msg("Could not ascertain agent registration status.")
	}
	log.Debug().Msgf("Registration status is %v", registered)

	// If the agent is not registered (or force registration requested) run a
	// registration flow
	if !registered || agent.Options.ForceRegister {
		agent.performRegistration(context.Background(), agent.Options.Server, agent.Options.Token, c)
		if agent.Options.ForceRegister {
			t.Reset()
		}
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
