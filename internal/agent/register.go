// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

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
func saveRegistration(cfg config.Config, r *api.RegistrationResponse, d api.DeviceInfo) {
	checkFatal := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("Could not save registration.")
		}
	}
	var providedHost string
	checkFatal(cfg.Get(config.PrefHost, &providedHost))

	if r.CloudhookURL != "" {
		checkFatal(cfg.Set(config.PrefCloudhookURL, r.CloudhookURL))
	}
	if r.RemoteUIURL != "" {
		checkFatal(cfg.Set(config.PrefRemoteUIURL, r.RemoteUIURL))
	}
	if r.Secret != "" {
		checkFatal(cfg.Set(config.PrefSecret, r.Secret))
	}
	if r.WebhookID != "" {
		checkFatal(cfg.Set(config.PrefWebhookID, r.WebhookID))
	}
	checkFatal(cfg.Set(config.PrefAPIURL, r.GenerateAPIURL(providedHost)))
	checkFatal(cfg.Set(config.PrefWebsocketURL, r.GenerateWebsocketURL(providedHost)))
	checkFatal(cfg.Set(config.PrefDeviceName, d.DeviceName()))
	checkFatal(cfg.Set(config.PrefDeviceID, d.DeviceID()))
	checkFatal(cfg.Set(config.PrefRegistered, true))
	checkFatal(cfg.Set(config.PrefVersion, config.AppVersion))
}

// performRegistration runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config.
func (agent *Agent) performRegistration(ctx context.Context, cfg config.Config) {
	log.Info().Msg("Registration required. Starting registration process.")
	if agent.options.Server != "" {
		if !validateRegistrationSetting("server", agent.options.Server) {
			log.Fatal().Msg("Server setting is not valid.")
		} else if err := cfg.Set(config.PrefHost, agent.options.Server); err != nil {
			log.Fatal().Err(err).Msg("Could not set host preference.")
		}
	}
	if agent.options.Token != "" {
		if !validateRegistrationSetting("token", agent.options.Token) {
			log.Fatal().Msg("Token setting is not valid.")
		} else if err := cfg.Set(config.PrefToken, agent.options.Token); err != nil {
			log.Fatal().Err(err).Msg("Could not set token preference.")
		}
	}

	device := newDevice(ctx)
	if !agent.options.Headless {
		userInputDone := make(chan struct{})
		agent.ui.DisplayRegistrationWindow(ctx, agent, cfg, userInputDone)
		<-userInputDone
	}
	resp, err := api.RegisterWithHass(ctx, cfg, device)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not register with Home Assistant.")
	}
	saveRegistration(cfg, resp, device)
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
	if !registered || agent.options.Register {
		agent.performRegistration(context.Background(), c)
		if agent.options.Register {
			t.Reset()
		}
	}
}

func validateRegistrationSetting(key, value string) bool {
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
