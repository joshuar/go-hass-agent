package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/mdns"
	"github.com/joshuar/go-hass-agent/internal/device"
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

	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			server := entry.AddrV4.String() + ":" + fmt.Sprint(entry.Port)
			log.Debug().Caller().Msgf("Found a server: %s", server)
			serverList.Append(server)
		}
	}()

	// Start the lookup
	mdns.Lookup("_home-assistant._tcp", entriesCh)
	close(entriesCh)

	if serverList == nil {
		log.Warn().Msg("Could not find any Home Assistant servers on the network")
	}
	// }
	// add http://localhost:8123 to the list of servers as a fall-back/default option
	serverList.Append("localhost:8123")
	return serverList
}

func (agent *Agent) getRegistrationHostInfo(ctx context.Context) *hass.RegistrationHost {

	msgPrinter := newMsgPrinter()

	registrationInfo := newRegistration()

	done := make(chan bool, 1)
	defer close(done)

	s := findServers()
	allServers, _ := s.Get()

	w := agent.App.NewWindow(msgPrinter.Sprintf("App Registration"))

	tokenSelect := widget.NewEntryWithData(registrationInfo.Token)

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		registrationInfo.Server.Set(s)
	})

	manualServerEntry := widget.NewEntryWithData(registrationInfo.Server)
	manualServerEntry.Validator = newHostPort()
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

	tlsSelect := widget.NewCheckWithData("", registrationInfo.UseTLS)

	form := widget.NewForm(
		widget.NewFormItem(msgPrinter.Sprintf("Token"), tokenSelect),
		widget.NewFormItem(msgPrinter.Sprintf("Auto-discovered Servers"), autoServerSelect),
		widget.NewFormItem(msgPrinter.Sprintf("Use Custom Server?"), manualServerSelect),
		widget.NewFormItem(msgPrinter.Sprintf("Manual Server Entry"), manualServerEntry),
		widget.NewFormItem(msgPrinter.Sprintf("Use TLS?"), tlsSelect),
	)
	form.OnSubmit = func() {
		s, _ := registrationInfo.Server.Get()
		log.Debug().Caller().
			Msgf("User selected server %s", s)

		w.Close()
		done <- true
	}
	form.OnCancel = func() {
		registrationInfo = nil
		done <- true
	}

	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(msgPrinter.Sprint("As an initial step, this app will need to log into your Home Assistant server and register itself.\nPlease enter the relevant details for your Home Assistant server url/port and a long-lived access token.")),
		form,
	))

	w.SetOnClosed(func() {
		done <- true
	})
	w.Show()
	<-done
	w.Close()
	return registrationInfo
}

func (agent *Agent) saveRegistration(r *hass.RegistrationResponse, h *hass.RegistrationHost) {
	host, _ := h.Server.Get()
	useTLS, _ := h.UseTLS.Get()
	agent.App.Preferences().SetString("Host", host)
	agent.App.Preferences().SetBool("UseTLS", useTLS)
	token, _ := h.Token.Get()
	agent.App.Preferences().SetString("Token", token)
	agent.App.Preferences().SetString("Version", agent.Version)
	if r.CloudhookURL != "" {
		agent.App.Preferences().SetString("CloudhookURL", r.CloudhookURL)
	}
	if r.RemoteUIURL != "" {
		agent.App.Preferences().SetString("RemoteUIURL", r.RemoteUIURL)
	}
	if r.Secret != "" {
		agent.App.Preferences().SetString("Secret", r.Secret)
	}
	if r.WebhookID != "" {
		agent.App.Preferences().SetString("WebhookID", r.WebhookID)
	}
}

func (agent *Agent) runRegistrationWorker(ctx context.Context) error {
	thisDevice := device.NewDevice()
	agent.App.Preferences().SetString("DeviceID", thisDevice.DeviceID())
	agent.App.Preferences().SetString("DeviceName", thisDevice.DeviceName())
	registrationHostInfo := agent.getRegistrationHostInfo(ctx)
	if registrationHostInfo != nil {
		registrationRequest := device.GenerateRegistrationRequest(thisDevice)
		appRegistrationInfo := hass.RegisterWithHass(registrationHostInfo, registrationRequest)
		if appRegistrationInfo != nil {
			agent.saveRegistration(appRegistrationInfo, registrationHostInfo)
			return nil
		} else {
			return errors.New("registration failed")
		}
	} else {
		return errors.New("problem getting registration information")
	}
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
