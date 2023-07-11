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
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	validate "github.com/go-playground/validator/v10"
)

// newRegistration creates a hass.RegistrationDetails object that contains
// information about both the Home Assistant server and the device running the
// agent needed to register the agent with Home Assistant.
func (agent *Agent) newRegistration(ctx context.Context, server, token string) *hass.RegistrationDetails {
	checkSet := func(value string, pref binding.String) {
		if err := pref.Set(value); err != nil {
			log.Warn().Err(err).
				Msgf("Could not set preference to provided value: %s", value)
		}
	}
	registrationInfo := &hass.RegistrationDetails{
		Server: binding.NewString(),
		Token:  binding.NewString(),
		Device: agent.setupDevice(ctx),
	}
	u, err := url.Parse(server)
	if err != nil {
		log.Warn().Err(err).
			Msg("Cannot parse provided URL. Ignoring")
	} else {
		checkSet(u.Host, registrationInfo.Server)
	}
	if token != "" {
		checkSet(token, registrationInfo.Token)
	}
	return registrationInfo
}

// registrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (agent *Agent) registrationWindow(ctx context.Context, registration *hass.RegistrationDetails, done chan struct{}) {
	s := findServers(ctx)
	allServers, _ := s.Get()

	w := agent.app.NewWindow(translator.Translate("App Registration"))

	tokenSelect := widget.NewEntryWithData(registration.Token)
	tokenSelect.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		if err := registration.Server.Set(s); err != nil {
			log.Debug().Err(err).
				Msg("Could not set server pref to selected value.")
		}
	})

	manualServerEntry := widget.NewEntryWithData(registration.Server)
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

	form := widget.NewForm(
		widget.NewFormItem(translator.Translate("Token"), tokenSelect),
		widget.NewFormItem(translator.Translate("Auto-discovered Servers"), autoServerSelect),
		widget.NewFormItem(translator.Translate("Use Custom Server?"), manualServerSelect),
		widget.NewFormItem(translator.Translate("Manual Server Entry"), manualServerEntry),
	)
	form.OnSubmit = func() {
		w.Close()
	}
	form.OnCancel = func() {
		registration = nil
		w.Close()
		ctx.Done()
	}

	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(
			translator.Translate(
				"As an initial step, this app will need to log into your Home Assistant server and register itself.\nPlease enter the relevant details for your Home Assistant server url/port and a long-lived access token.")),
		form,
	))

	w.SetOnClosed(func() {
		registration = nil
		close(done)
	})

	w.Show()
	w.Close()
}

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func (agent *Agent) saveRegistration(r *hass.RegistrationResponse, h *hass.RegistrationDetails) {
	providedHost, _ := h.Server.Get()
	hostURL, _ := url.Parse(providedHost)
	agent.SetPref("Host", hostURL.String())

	token, _ := h.Token.Get()
	agent.SetPref("Token", token)

	if r.CloudhookURL != "" {
		agent.SetPref("CloudhookURL", r.CloudhookURL)
	}
	if r.RemoteUIURL != "" {
		agent.SetPref("RemoteUIURL", r.RemoteUIURL)
	}
	if r.Secret != "" {
		agent.SetPref("Secret", r.Secret)
	}
	if r.WebhookID != "" {
		agent.SetPref("WebhookID", r.WebhookID)
	}
	agent.SetPref("ApiURL", r.GenerateAPIURL(providedHost))
	agent.SetPref("WebSocketURL", r.GenerateWebsocketURL(providedHost))

	agent.SetPref("DeviceName", h.Device.DeviceName())
	agent.SetPref("DeviceID", h.Device.AppID())

	agent.SetPref("Registered", true)

	agent.SetPref("Version", agent.Version)

	registryPath, err := agent.extraStoragePath("sensorRegistry")
	if err != nil {
		return
	} else {
		if err := os.RemoveAll(registryPath.Path()); err != nil {
			log.Debug().Err(err).Msg("Could not remove existing registry DB.")
		}
	}

	// ! https://github.com/fyne-io/fyne/issues/3170
	time.Sleep(110 * time.Millisecond)
}

// registerWithUI handles a registration flow via a graphical interface
func (agent *Agent) registerWithUI(ctx context.Context, registration *hass.RegistrationDetails) (*hass.RegistrationResponse, error) {
	done := make(chan struct{})
	agent.registrationWindow(ctx, registration, done)
	<-done
	if !registration.Validate() {
		return nil, errors.New("registration details not complete")
	}
	return hass.RegisterWithHass(registration)
}

// registerWithoutUI handles a registration flow without any graphical interface
// (using values provided via the command-line).
func (agent *Agent) registerWithoutUI(ctx context.Context, registration *hass.RegistrationDetails) (*hass.RegistrationResponse, error) {
	if !registration.Validate() {
		log.Debug().Msg("Registration details not complete.")
		return nil, errors.New("registration details not complete")
	}
	return hass.RegisterWithHass(registration)
}

func (agent *Agent) registrationProcess(ctx context.Context, server, token string, force, headless bool, done chan struct{}) {
	appConfig := agent.LoadConfig()
	// If the agent isn't registered but the config is valid, set the agent as
	// registered and continue execution. Required check for versions upgraded
	// from v1.2.6 and below.
	if !agent.IsRegistered() && appConfig.Validate() == nil {
		appConfig.prefs.SetBool("Registered", true)
		close(done)
	}
	// If the app is not registered, run a registration flow
	if !agent.IsRegistered() || force {
		log.Info().Msg("Registration required. Starting registration process.")
		// The app is registered, continue (config check performed later).

		registration := agent.newRegistration(ctx, server, token)
		var registrationResponse *hass.RegistrationResponse
		var err error
		if headless {
			registrationResponse, err = agent.registerWithoutUI(ctx, registration)
		} else {
			registrationResponse, err = agent.registerWithUI(ctx, registration)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("Could not register device with Home Assistant.")
		}

		agent.saveRegistration(registrationResponse, registration)
	}

	close(done)
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

// hostValidator is a custom fyne validator that will validate a string is a
// valid hostname:port combination
func hostValidator() fyne.StringValidator {
	v := validate.New()
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
