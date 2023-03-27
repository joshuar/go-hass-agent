package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	validate "github.com/go-playground/validator/v10"
)

const (
	HelpText = ""
)

func newRegistration() *hass.RegistrationHost {
	return &hass.RegistrationHost{
		Server: binding.NewString(),
		Token:  binding.NewString(),
		UseTLS: binding.NewBool(),
	}
}

func findServers() binding.StringList {

	serverList := binding.NewStringList()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Warn().Msgf("Failed to initialize resolver:", err.Error())
	} else {
		entries := make(chan *zeroconf.ServiceEntry)
		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				server := entry.AddrIPv4[0].String() + ":" + fmt.Sprint(entry.Port)
				serverList.Append(server)
				log.Debug().Caller().
					Msgf("Found a record %s", server)
			}
		}(entries)

		log.Info().Msg("Looking for Home Assistant instances on the network...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err = resolver.Browse(ctx, "_home-assistant._tcp", "local.", entries)
		if err != nil {
			log.Warn().Msgf("Failed to browse:", err.Error())
		}

		<-ctx.Done()
		if serverList == nil {
			log.Warn().Msg("Could not find any Home Assistant servers on the network")
		}
	}
	// add http://localhost:8123 to the list of servers as a fall-back/default option
	serverList.Append("localhost:8123")
	return serverList
}

func (agent *Agent) getRegistrationHostInfo() *hass.RegistrationHost {

	registrationInfo := newRegistration()

	done := make(chan bool, 1)

	s := findServers()
	allServers, _ := s.Get()

	w := agent.App.NewWindow(agent.MsgPrinter.Sprintf("App Registration"))

	serverSelect := widget.NewSelect(allServers, func(s string) {
		registrationInfo.Server.Set(s)
	})
	serverManual := widget.NewEntryWithData(registrationInfo.Server)
	serverManual.Validator = newHostPort()
	serverManual.Disable()
	manualServerSelect := widget.NewCheck(agent.MsgPrinter.Sprintf("Use Custom Server"), func(b bool) {
		switch b {
		case true:
			serverManual.Enable()
			serverSelect.Disable()
		case false:
			serverManual.Disable()
			serverSelect.Enable()
		}
	})
	tokenSelect := widget.NewEntryWithData(registrationInfo.Token)
	// tokenSelect.Validator = validation.NewRegexp(`^[A-Za-z0-9_-\.]+$`, "token can only contain letters, numbers, '_', '-' and '.'")
	tlsSelect := widget.NewCheckWithData(agent.MsgPrinter.Sprintf("Use TLS?"), registrationInfo.UseTLS)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: agent.MsgPrinter.Sprintf("Token"), Widget: tokenSelect},
			{Text: agent.MsgPrinter.Sprintf("Found Server"), Widget: serverSelect},
			{Text: agent.MsgPrinter.Sprintf("Manual Server"), Widget: container.NewHBox(manualServerSelect, serverManual)},
			{Widget: tlsSelect},
		},
		OnSubmit: func() { // optional, handle form submission
			s, _ := registrationInfo.Server.Get()
			log.Debug().Caller().
				Msgf("User selected server %s", s)
			done <- true
		},
	}

	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(agent.MsgPrinter.Sprint("As an initial step, this app will need to log into your Home Assistant server and register itself.\nPlease enter the relevant details for your Home Assistant server url/port and a long-lived access token.")),
		form,
	))

	w.Show()
	<-done
	w.Close()
	return registrationInfo
}

// newHostPort is a custom fyne validator that will validate a string is a
// valid hostname:port combination
func newHostPort() fyne.StringValidator {
	v := validate.New()
	return func(text string) error {
		if err := v.Var(text, "hostname_port"); err != nil {
			return errors.New("you need to specify a valid hostname:port combination")
		}
		return nil
	}
}

func (a *Agent) saveRegistration(r *hass.RegistrationResponse, h *hass.RegistrationHost) {
	host, _ := h.Server.Get()
	useTLS, _ := h.UseTLS.Get()
	a.App.Preferences().SetString("Host", host)
	a.App.Preferences().SetBool("UseTLS", useTLS)
	token, _ := h.Token.Get()
	a.App.Preferences().SetString("Token", token)
	a.App.Preferences().SetString("Version", a.Version)
	if r.CloudhookURL != "" {
		a.App.Preferences().SetString("CloudhookURL", r.CloudhookURL)
	}
	if r.RemoteUIURL != "" {
		a.App.Preferences().SetString("RemoteUIURL", r.RemoteUIURL)
	}
	if r.Secret != "" {
		a.App.Preferences().SetString("Secret", r.Secret)
	}
	if r.WebhookID != "" {
		a.App.Preferences().SetString("WebhookID", r.WebhookID)
	}
}

func (a *Agent) runRegistrationWorker() {
	device := hass.NewDevice()
	registrationHostInfo := a.getRegistrationHostInfo()
	registrationRequest := hass.GenerateRegistrationRequest(device)
	appRegistrationInfo := hass.RegisterWithHass(registrationHostInfo, registrationRequest)
	if appRegistrationInfo != nil {
		a.saveRegistration(appRegistrationInfo, registrationHostInfo)
	}
}
