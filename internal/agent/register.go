package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/joshuar/go-hass-agent/internal/hass"
	log "github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	validate "github.com/go-playground/validator/v10"
)

const (
	HelpText = "As an initial step, this app will need to log into your Home Assistant server and register itself.\nPlease enter the relevant details for your Home Assistant server url/port and a long-lived access token."
)

type RegistrationInfo struct {
	Server, Token binding.String
	UseTLS        binding.Bool
}

func NewRegistration() *RegistrationInfo {
	return &RegistrationInfo{
		Server: binding.NewString(),
		Token:  binding.NewString(),
		UseTLS: binding.NewBool(),
	}
}

func findServers() binding.StringList {

	serverList := binding.NewStringList()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			server := entry.AddrIPv4[0].String() + ":" + fmt.Sprint(entry.Port)
			serverList.Append(server)
			log.Debugf("Found a record %s", server)
		}
	}(entries)

	log.Info("Looking for Home Assistant instances on the network...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = resolver.Browse(ctx, "_home-assistant._tcp", "local.", entries)
	if err != nil {
		log.Fatalf("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	if serverList == nil {
		log.Warn("Could not find any Home Assistant servers on the network")
	}
	// add http://localhost:8123 to the list of servers as a fall-back/default option
	serverList.Append("localhost:8123")
	return serverList
}

func (agent *Agent) GetRegistrationInfo() *RegistrationInfo {

	registrationInfo := NewRegistration()

	done := make(chan bool, 1)

	s := findServers()
	allServers, _ := s.Get()

	log.Debug("Prompt for registration details")

	w := agent.App.NewWindow("App Registration")

	serverSelect := widget.NewSelect(allServers, func(s string) {
		registrationInfo.Server.Set(s)
	})
	serverManual := widget.NewEntryWithData(registrationInfo.Server)
	serverManual.Validator = NewHostPort()
	serverManual.Disable()
	manualServerSelect := widget.NewCheck("Use Custom Server", func(b bool) {
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
	tokenSelect.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "token can only contain letters, numbers, '_', and '-'")
	tlsSelect := widget.NewCheckWithData("Use TLS?", registrationInfo.UseTLS)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Token", Widget: tokenSelect},
			{Text: "Found Server", Widget: serverSelect},
			{Text: "Manual Server", Widget: container.NewHBox(manualServerSelect, serverManual)},
			{Widget: tlsSelect},
		},
		OnSubmit: func() { // optional, handle form submission
			s, _ := registrationInfo.Server.Get()
			t, _ := registrationInfo.Token.Get()
			log.Debugf("User selected server %s and token %s", s, t)
			done <- true
		},
	}

	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(HelpText),
		form,
	))

	w.Show()
	<-done
	w.Close()
	return registrationInfo
}

// NewHostPort is a custom fyne validator that will validate a string is a
// valid hostname:port combination
func NewHostPort() fyne.StringValidator {
	v := validate.New()
	return func(text string) error {
		if err := v.Var(text, "hostname_port"); err != nil {
			return errors.New("you need to specify a valid hostname:port combination")
		}
		return nil
	}
}

func (a *Agent) SaveRegistration(r *hass.RegistrationResponse) error {
	a.App.Preferences().SetString("CloudhookURL", r.CloudhookURL)
	a.App.Preferences().SetString("RemoteUIURL", r.RemoteUIURL)
	a.App.Preferences().SetString("Secret", r.Secret)
	a.App.Preferences().SetString("WebhookID", r.WebhookID)
	return nil
}

func (a *Agent) RegisterWithHass(r *RegistrationInfo) *hass.RegistrationResponse {
	return &hass.RegistrationResponse{}
}
