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
	"github.com/go-playground/validator/v10"
)

const (
	explainRegistration = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`
)

type RegistrationDetails struct {
	serverBinding, tokenBinding binding.String
}

func (r *RegistrationDetails) Server() string {
	var s string
	var err error
	if s, err = r.serverBinding.Get(); err != nil {
		log.Warn().Err(err).Msg("Unable to retrieve server from registration details.")
		return ""
	}
	return s
}

func (r *RegistrationDetails) Token() string {
	var s string
	var err error
	if s, err = r.tokenBinding.Get(); err != nil {
		log.Warn().Err(err).Msg("Unable to retrieve token from registration details.")
		return ""
	}
	return s
}

func (r *RegistrationDetails) Validate() bool {
	validate := validator.New()
	check := func(value string, validation string) bool {
		if err := validate.Var(value, validation); err != nil {
			return false
		}
		return true
	}
	if server, _ := r.serverBinding.Get(); !check(server, "required,http_url") {
		return false
	}
	if token, _ := r.tokenBinding.Get(); !check(token, "required") {
		return false
	}
	return true
}

// newRegistration creates a hass.RegistrationDetails object that contains
// information about both the Home Assistant server and the device running the
// agent needed to register the agent with Home Assistant.
func newRegistration(server, token string) *RegistrationDetails {
	checkSet := func(value string, pref binding.String) {
		if err := pref.Set(value); err != nil {
			log.Warn().Err(err).
				Msgf("Could not set preference to provided value: %s", value)
		}
	}
	registrationInfo := &RegistrationDetails{
		serverBinding: binding.NewString(),
		tokenBinding:  binding.NewString(),
	}
	u, err := url.Parse(server)
	if err != nil {
		log.Warn().Err(err).
			Msg("Cannot parse provided URL. Ignoring")
	} else {
		checkSet(u.Host, registrationInfo.serverBinding)
	}
	if token != "" {
		checkSet(token, registrationInfo.tokenBinding)
	}
	return registrationInfo
}

// registrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (agent *Agent) registrationWindow(ctx context.Context, registration *RegistrationDetails, done chan struct{}) {
	s := findServers(ctx)
	allServers, _ := s.Get()

	agent.mainWindow.SetTitle(translator.Translate("App Registration"))
	// w := agent.app.NewWindow()

	tokenSelect := widget.NewEntryWithData(registration.tokenBinding)
	tokenSelect.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		if err := registration.serverBinding.Set(s); err != nil {
			log.Debug().Err(err).
				Msg("Could not set server pref to selected value.")
		}
	})

	manualServerEntry := widget.NewEntryWithData(registration.serverBinding)
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
		agent.mainWindow.Hide()
		close(done)
	}
	form.OnCancel = func() {
		log.Warn().Msg("Cancelling registration.")
		close(done)
		registration = nil
		agent.mainWindow.Close()
		ctx.Done()
	}

	agent.mainWindow.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(
			translator.Translate(explainRegistration)),
		form,
	))

	agent.mainWindow.SetOnClosed(func() {
		log.Debug().Msg("Closed")
		registration = nil
		close(done)
	})

	agent.mainWindow.Show()
}

// saveRegistration stores the relevant information from the registration
// request and the successful response in the agent preferences. This includes,
// most importantly, details on the URL that should be used to send subsequent
// requests to Home Assistant.
func (agent *Agent) saveRegistration(r *hass.RegistrationResponse, h *RegistrationDetails, c config, d hass.DeviceInfo) {
	providedHost, _ := h.serverBinding.Get()
	hostURL, _ := url.Parse(providedHost)
	c.Set("Host", hostURL.String())

	token, _ := h.tokenBinding.Get()
	c.Set(PrefToken, token)

	if r.CloudhookURL != "" {
		c.Set("CloudhookURL", r.CloudhookURL)
	}
	if r.RemoteUIURL != "" {
		c.Set("RemoteUIURL", r.RemoteUIURL)
	}
	if r.Secret != "" {
		c.Set(PrefSecret, r.Secret)
	}
	if r.WebhookID != "" {
		c.Set(PrefWebhookID, r.WebhookID)
	}
	c.Set(PrefApiURL, r.GenerateAPIURL(providedHost))
	c.Set(PrefWebsocketURL, r.GenerateWebsocketURL(providedHost))

	c.Set("DeviceName", d.DeviceName())
	c.Set("DeviceID", d.DeviceID())

	agent.SetRegistered(true)

	c.Set("Version", agent.Version)

	registryPath, err := extraStoragePath("sensorRegistry")
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

func (agent *Agent) registrationProcess(ctx context.Context, server, token string, force, headless bool, done chan struct{}) {
	appConfig := agent.LoadConfig()
	// If the agent isn't registered but the config is valid, set the agent as
	// registered and continue execution. Required check for versions upgraded
	// from v1.2.6 and below.
	if !agent.IsRegistered() && ValidateConfig(appConfig) == nil {
		agent.SetRegistered(true)
		close(done)
	}
	// If the app is not registered, run a registration flow
	if !agent.IsRegistered() || force {
		log.Info().Msg("Registration required. Starting registration process.")
		// agent.showFirstRunWindow(ctx)
		// The app is registered, continue (config check performed later).

		registration := newRegistration(server, token)
		device := agent.setupDevice(ctx)
		if !headless {
			done := make(chan struct{})
			agent.registrationWindow(ctx, registration, done)
			<-done
		}
		if !registration.Validate() {
			log.Fatal().Msg("Registration details not valid.")
		}
		registrationResponse, err := hass.RegisterWithHass(ctx, registration, device)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not register with Home Assistant.")
		}
		agent.saveRegistration(registrationResponse, registration, appConfig, device)
		log.Info().Msg("Successfully registered agent.")
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
