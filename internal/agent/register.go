// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-playground/validator/v10"
)

const (
	explainRegistration = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`
)

// registrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (agent *Agent) registrationWindow(ctx context.Context, done chan struct{}) {
	agent.mainWindow.SetTitle(translator.Translate("App Registration"))

	var allFormItems []*widget.FormItem

	allFormItems = append(allFormItems, agent.serverConfigItems(ctx)...)
	// allFormItems = append(allFormItems, agent.mqttConfigItems()...)
	registrationForm := widget.NewForm(allFormItems...)
	registrationForm.OnSubmit = func() {
		agent.mainWindow.Hide()
		close(done)
	}
	registrationForm.OnCancel = func() {
		log.Warn().Msg("Cancelling registration.")
		close(done)
		agent.mainWindow.Close()
		ctx.Done()
	}

	agent.mainWindow.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(translator.Translate(explainRegistration)),
		registrationForm,
	))

	agent.mainWindow.SetOnClosed(func() {
		log.Debug().Msg("Closed")
	})

	agent.mainWindow.Show()
}

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
	checkFatal(agent.Config.Get(config.PrefHost, &providedHost))
	// hostURL, _ := url.Parse(providedHost)
	// checkFatal(agent.Config.Set(config.PrefHost, hostURL.String()))

	if r.CloudhookURL != "" {
		checkFatal(agent.Config.Set(config.PrefCloudhookURL, r.CloudhookURL))
	}
	if r.RemoteUIURL != "" {
		checkFatal(agent.Config.Set(config.PrefRemoteUIURL, r.RemoteUIURL))
	}
	if r.Secret != "" {
		checkFatal(agent.Config.Set(config.PrefSecret, r.Secret))
	}
	if r.WebhookID != "" {
		checkFatal(agent.Config.Set(config.PrefWebhookID, r.WebhookID))
	}
	checkFatal(agent.Config.Set(config.PrefApiURL, r.GenerateAPIURL(providedHost)))
	checkFatal(agent.Config.Set(config.PrefWebsocketURL, r.GenerateWebsocketURL(providedHost)))
	checkFatal(agent.Config.Set(config.PrefDeviceName, d.DeviceName()))
	checkFatal(agent.Config.Set(config.PrefDeviceID, d.DeviceID()))
	checkFatal(agent.Config.Set(config.PrefRegistered, true))
	checkFatal(agent.Config.Set(config.PrefVersion, agent.Version))

	registryPath, err := agent.Config.StoragePath("sensorRegistry")
	if err != nil {
		return
	} else {
		if err := os.RemoveAll(registryPath); err != nil {
			log.Debug().Err(err).Msg("Could not remove existing registry DB.")
		}
	}

	// ! https://github.com/fyne-io/fyne/issues/3170
	time.Sleep(110 * time.Millisecond)
}

// registrationProcess runs through a registration flow. If the agent is already
// registered, it will exit unless the force parameter is true. Otherwise, it
// will action a registration workflow displaying a GUI for user input of
// registration details and save the results into the agent config
func (agent *Agent) registrationProcess(ctx context.Context, server, token string, force, headless bool, done chan struct{}) {
	// If the agent isn't registered but the config is valid, set the agent as
	// registered and continue execution. Required check for versions upgraded
	// from v1.2.6 and below.
	if !agent.IsRegistered() && ValidateConfig(agent.Config) == nil {
		agent.SetRegistered(true)
		close(done)
	}
	// If the app is not registered, run a registration flow
	if !agent.IsRegistered() || force {
		log.Info().Msg("Registration required. Starting registration process.")
		if server != "" {
			if !validateRegistrationSetting("server", server) {
				log.Fatal().Msg("Server setting is not valid.")
			}
		} else {
			if err := agent.Config.Set(config.PrefHost, token); err != nil {
				log.Fatal().Err(err).Msg("Could not set host preference.")
			}
		}
		if token != "" {
			if !validateRegistrationSetting("token", token) {
				log.Fatal().Msg("Token setting is not valid.")
			}
		} else {
			if err := agent.Config.Set(config.PrefToken, token); err != nil {
				log.Fatal().Err(err).Msg("Could not set token preference.")
			}
		}

		device := agent.setupDevice(ctx)
		if !headless {
			userInputDone := make(chan struct{})
			agent.registrationWindow(ctx, userInputDone)
			<-userInputDone
		}
		registrationResponse, err := api.RegisterWithHass(ctx, agent.Config, device)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not register with Home Assistant.")
		}
		agent.saveRegistration(registrationResponse, device)
		log.Info().Msg("Successfully registered agent.")
	}

	close(done)
}

// serverSelectionForm generates a fyne.CanvasObject consisting of a form for
// selecting a server to register the agent against
func (agent *Agent) serverConfigItems(ctx context.Context) []*widget.FormItem {
	s := findServers(ctx)
	allServers, _ := s.Get()

	token := binding.BindPreferenceString(config.PrefToken, agent.app.Preferences())
	server := binding.BindPreferenceString(config.PrefHost, agent.app.Preferences())

	tokenSelect := widget.NewEntryWithData(token)
	tokenSelect.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		if err := server.Set(s); err != nil {
			log.Debug().Err(err).
				Msg("Could not set server pref to selected value.")
		}
	})

	manualServerEntry := widget.NewEntryWithData(server)
	manualServerEntry.Validator = hostValidator()
	manualServerEntry.Disable()
	manualServerSelect := widget.NewCheck("", func(b bool) {
		switch b {
		case true:
			manualServerEntry.Enable()
			autoServerSelect.Disable()
		case false:
			manualServerEntry.Disable()
			autoServerSelect.Enable()
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(translator.Translate("Token"), tokenSelect),
		widget.NewFormItem(translator.Translate("Auto-discovered Servers"), autoServerSelect),
		widget.NewFormItem(translator.Translate("Use Custom Server?"), manualServerSelect),
		widget.NewFormItem(translator.Translate("Manual Server Entry"), manualServerEntry))

	return items
}

// mqttConfigForm returns a fyne.CanvasObject consisting of a form for
// configuring the agent to use an MQTT for pub/sub functionality
func (agent *Agent) mqttConfigItems() []*widget.FormItem {

	mqttServer := binding.BindPreferenceString(config.PrefMQTTServer, agent.app.Preferences())

	mqttServerEntry := widget.NewEntryWithData(mqttServer)
	mqttServerEntry.Validator = hostValidator()
	mqttServerEntry.Disable()

	mqttEnabled := widget.NewCheck("", func(b bool) {
		switch b {
		case true:
			mqttServerEntry.Enable()
		case false:
			mqttServerEntry.Disable()
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(translator.Translate("Use MQTT?"), mqttEnabled),
		widget.NewFormItem(translator.Translate("MQTT Server"), mqttServerEntry))

	return items
}

// findServers is a helper function to generate a list of Home Assistant servers
// via local network auto-discovery.
func findServers(ctx context.Context) binding.StringList {
	serverList := binding.NewStringList()

	// add http://localhost:8123 to the list of servers as a fall-back/default
	// option
	if err := serverList.Append("http://localhost:8123"); err != nil {
		log.Debug().Err(err).
			Msg("Unable to set a default server.")
	}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize resolver.")
	} else {
		entries := make(chan *zeroconf.ServiceEntry)
		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				var server string
				for _, t := range entry.Text {
					if value, found := strings.CutPrefix(t, "base_url="); found {
						server = value
					}
				}
				if server != "" {
					if err := serverList.Append(server); err != nil {
						log.Warn().Err(err).
							Msgf("Unable to add found server %s to server list.", server)
					}
				} else {
					log.Debug().Msgf("Entry %s did not have a base_url value. Not using it.", entry.HostName)
				}
			}
		}(entries)

		log.Info().Msg("Looking for Home Assistant instances on the network...")
		searchCtx, searchCancel := context.WithTimeout(ctx, time.Second*5)
		defer searchCancel()
		err = resolver.Browse(searchCtx, "_home-assistant._tcp", "local.", entries)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to browse")
		}

		<-searchCtx.Done()
	}
	return serverList
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

// hostValidator is a custom fyne validator that will validate a string is a
// valid hostname:port combination
func hostValidator() fyne.StringValidator {
	v := validator.New()
	return func(text string) error {
		if v.Var(text, "http_url") != nil {
			return errors.New("you need to specify a valid url")
		}
		if _, err := url.Parse(text); err != nil {
			return errors.New("url is invalid")
		}
		return nil
	}
}
