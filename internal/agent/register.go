// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"os"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"

	"github.com/go-playground/validator/v10"
)

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func (agent *Agent) saveRegistration(r *api.RegistrationResponse, d api.DeviceInfo) {
	checkFatal := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("Could not save registration.")
		}
	}
	var providedHost string
	checkFatal(agent.config.Get(config.PrefHost, &providedHost))

	if r.CloudhookURL != "" {
		checkFatal(agent.config.Set(config.PrefCloudhookURL, r.CloudhookURL))
	}
	if r.RemoteUIURL != "" {
		checkFatal(agent.config.Set(config.PrefRemoteUIURL, r.RemoteUIURL))
	}
	if r.Secret != "" {
		checkFatal(agent.config.Set(config.PrefSecret, r.Secret))
	}
	if r.WebhookID != "" {
		checkFatal(agent.config.Set(config.PrefWebhookID, r.WebhookID))
	}
	checkFatal(agent.config.Set(config.PrefAPIURL, r.GenerateAPIURL(providedHost)))
	checkFatal(agent.config.Set(config.PrefWebsocketURL, r.GenerateWebsocketURL(providedHost)))
	checkFatal(agent.config.Set(config.PrefDeviceName, d.DeviceName()))
	checkFatal(agent.config.Set(config.PrefDeviceID, d.DeviceID()))
	checkFatal(agent.config.Set(config.PrefRegistered, true))
	checkFatal(agent.config.Set(config.PrefVersion, agent.version))

	registryPath, err := agent.config.StoragePath("sensorRegistry")
	if err != nil {
		return
	} else {
		if err := os.RemoveAll(registryPath); err != nil {
			log.Debug().Err(err).Msg("Could not remove existing registry DB.")
		}
	}
}

// registrationProcess runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config
func (agent *Agent) registrationProcess(ctx context.Context, server, token string, force, headless bool, done chan struct{}) {
	var registered bool
	if err := agent.config.Get(config.PrefRegistered, &registered); err != nil {
		log.Fatal().Err(err).Msg("Could not ascertain agent registration status.")
	}
	log.Debug().Msgf("Registration status is %v", registered)
	// If the config is valid, but the agent is not registered, set the agent as
	// registered and continue execution. Required check for versions upgraded
	// from v1.2.6 and below.
	if ValidateConfig(agent.config) == nil {
		if !registered {
			if err := agent.config.Set(config.PrefRegistered, true); err != nil {
				log.Fatal().Err(err).Msg("Could not set registered status.")
			}
			close(done)
		}
	}
	// If the agent is not registered (or force registration requested) run a
	// registration flow
	if !registered || force {
		log.Info().Msg("Registration required. Starting registration process.")
		if server != "" {
			if !validateRegistrationSetting("server", server) {
				log.Fatal().Msg("Server setting is not valid.")
			}
		} else {
			if err := agent.config.Set(config.PrefHost, token); err != nil {
				log.Fatal().Err(err).Msg("Could not set host preference.")
			}
		}
		if token != "" {
			if !validateRegistrationSetting("token", token) {
				log.Fatal().Msg("Token setting is not valid.")
			}
		} else {
			if err := agent.config.Set(config.PrefToken, token); err != nil {
				log.Fatal().Err(err).Msg("Could not set token preference.")
			}
		}

		device := agent.setupDevice(ctx)
		if !headless {
			userInputDone := make(chan struct{})
			agent.ui.DisplayRegistrationWindow(ctx, userInputDone)
			<-userInputDone
		}
		registrationResponse, err := api.RegisterWithHass(ctx, agent, device)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not register with Home Assistant.")
		}
		agent.saveRegistration(registrationResponse, device)
		log.Info().Msg("Successfully registered agent.")
	}

	close(done)
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
