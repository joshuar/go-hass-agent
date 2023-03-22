package agent

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

type Agent struct {
	App           fyne.App
	Tray          fyne.Window
	config        AppConfig
	hassConfig    *hass.ConfigResponse
	Name, Version string
}

func NewAgent() *Agent {
	a := NewUI()
	return &Agent{
		App:     a,
		Name:    Name,
		Version: Version,
	}
}

func (a *Agent) LoadConfig(done chan bool) {
	// go func() {
	for {
		CloudhookURL := a.App.Preferences().String("CloudhookURL")
		RemoteUIURL := a.App.Preferences().String("RemoteUIURL")
		WebhookID := a.App.Preferences().String("WebhookID")
		InstanceURL := a.App.Preferences().String("InstanceURL")

		a.config.secret = a.App.Preferences().String("Secret")
		a.config.token = a.App.Preferences().String("Token")

		switch {
		case CloudhookURL != "":
			a.config.RestAPIURL = CloudhookURL
			log.Debug().Caller().
				Msgf("Using CloudhookURL %s for Home Assistant access", a.config.RestAPIURL)
			done <- true
			return
		case RemoteUIURL != "" && WebhookID != "":
			a.config.RestAPIURL = RemoteUIURL + "/api/webhook/" + WebhookID
			log.Debug().Caller().
				Msgf("Using RemoteUIURL %s for Home Assistant access", a.config.RestAPIURL)
			done <- true
			return
		case WebhookID != "" && InstanceURL != "":
			a.config.RestAPIURL = InstanceURL + "/api/webhook/" + WebhookID
			log.Debug().Caller().
				Msgf("Using InstanceURL %s for Home Assistant access", a.config.RestAPIURL)
			done <- true
			return
		default:
			log.Warn().Msg("No suitable existing config found! Starting new registration process")
			device := hass.NewDevice()
			registrationHostInfo := a.GetRegistrationHostInfo()
			registrationRequest := hass.GenerateRegistrationRequest(device)
			appRegistrationInfo := hass.RegisterWithHass(registrationHostInfo, registrationRequest)
			a.SaveRegistration(appRegistrationInfo, registrationHostInfo)
		}
	}
	// }()
}

func (a *Agent) GetConfigVersion() string {
	return a.App.Preferences().String("Version")
}

func (a *Agent) GetHassConfig(configLoaded chan bool) {
	<-configLoaded
	a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
}

func (a *Agent) SetupSystemTray() {
	configLoaded := make(chan bool, 1)
	a.LoadConfig(configLoaded)
	<-configLoaded
	a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
	a.Tray = a.App.NewWindow("System Tray")
	a.Tray.SetMaster()
	if desk, ok := a.App.(desktop.App); ok {
		log.Debug().Caller().
			Msg("Config loaded successfully. Creating tray icon.")

		menuItemAbout := fyne.NewMenuItem("About", func() {
			w := a.App.NewWindow("About " + a.Name)
			w.SetContent(container.New(layout.NewVBoxLayout(),
				widget.NewLabel("App Version: "+a.Version),
				widget.NewLabel("Home Assistant Version: "+a.hassConfig.Version),
				widget.NewButton("Ok", func() {
					w.Close()
				}),
			))
			w.Show()
		})
		menu := fyne.NewMenu(a.Name, menuItemAbout)
		desk.SetSystemTrayMenu(menu)
	}
	a.Tray.Hide()
}

func (a *Agent) SetupRunners() {
	go hass.RunLocationUpdater()
}
