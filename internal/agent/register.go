package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	log "github.com/sirupsen/logrus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	validate "github.com/go-playground/validator/v10"
)

const (
	HelpText = "As an initial step, this app will need to log into your Home Assistant server and register itself.\nPlease enter the relevant details for your Home Assistant server url/port and a long-lived access token."
)

type RegistrationInfo struct {
	Server, Token string
	UseTLS        bool
}

func (r *RegistrationInfo) IsValid() bool {
	if r.Server == "" {
		return false
	}
	if r.Token == "" {
		return false
	}
	return true
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
		log.Fatalln("Failed to browse:", err.Error())
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

	registrationInfo := &RegistrationInfo{
		UseTLS: true,
	}

	s := findServers()
	allServers, _ := s.Get()

	w := agent.ui.NewWindow("App Registration")

	customServer := make(chan bool)
	serverSelect := widget.NewSelect(allServers, func(s string) {
		log.Debugf("User selected server %s", s)
		registrationInfo.Server = s
	})
	manualServerSelect := widget.NewCheck("Use Custom Server", func(b bool) {
		log.Debugf("Use custom server %v", b)
		customServer <- b
	})

	serverManual := widget.NewEntry()
	serverManual.Validator = NewHostPort()
	tokenSelect := widget.NewEntry()
	tokenSelect.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "token can only contain letters, numbers, '_', and '-'")

	defaultItems := []*widget.FormItem{
		{Text: "Token", Widget: tokenSelect},
		{Text: "Server", Widget: container.New(layout.NewHBoxLayout(),
			serverSelect,
			widget.NewCheck("Use TLS?", func(b bool) {
				switch b {
				case true:
					registrationInfo.UseTLS = true
				case false:
					registrationInfo.UseTLS = false
				}
			}),
		)},
		{Widget: manualServerSelect},
	}

	form := &widget.Form{
		Items: defaultItems,
		OnSubmit: func() { // optional, handle form submission
			if manualServerSelect.Checked {
				registrationInfo.Server = serverManual.Text
			}
			registrationInfo.Token = tokenSelect.Text
			if registrationInfo.IsValid() {
				w.Close()
			} else {
				err := errors.New("you need to specify both a token and server")
				dialog.ShowError(err, w)
			}
		},
	}
	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(HelpText),
		form,
	))

	go func(f *widget.Form) {
		for useCustomServer := range customServer {
			if useCustomServer {
				if len(f.Items) == 3 {
					f.Append("Manual Server:", serverManual)
					f.Refresh()
				}
			} else {
				registrationInfo.Server = serverSelect.Selected
				if len(f.Items) != 3 {
					f.Items = f.Items[:3]
					f.Refresh()
				}
			}
		}
	}(form)

	w.ShowAndRun()
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
