package agent

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/go-hass-agent/internal/hass"
	log "github.com/sirupsen/logrus"
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
			log.Debugf("Using CloudhookURL %s for REST API access", a.config.RestAPIURL)
			done <- true
			return
		case RemoteUIURL != "" && WebhookID != "":
			a.config.RestAPIURL = RemoteUIURL + "/api/webhook/" + WebhookID
			log.Debugf("Using RemoteUIURL %s for REST API access", a.config.RestAPIURL)
			done <- true
			return
		case WebhookID != "" && InstanceURL != "":
			a.config.RestAPIURL = InstanceURL + "/api/webhook/" + WebhookID
			log.Debugf("Using InstanceURL %s for REST API access", a.config.RestAPIURL)
			done <- true
			return
		default:
			log.Warn("No suitable existing config found, running registration")
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

func (a *Agent) SetupSystemTray(configLoaded chan bool) {
	a.LoadConfig(configLoaded)
	<-configLoaded
	a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
	a.Tray = a.App.NewWindow("System Tray")
	a.Tray.SetMaster()
	if desk, ok := a.App.(desktop.App); ok {
		log.Debug("Creating tray icon")

		ha_version := fyne.NewMenuItem("Home Assistant Version", func() {
			w := a.App.NewWindow("Home Assistant Version")
			w.SetContent(widget.NewLabel(a.hassConfig.Version))
			w.Show()
		})
		app_version := fyne.NewMenuItem("App Version", func() {
			w := a.App.NewWindow("App Version")
			w.SetContent(widget.NewLabel(a.Version))
			w.Show()
		})
		menu := fyne.NewMenu(a.Name, ha_version, app_version)
		desk.SetSystemTrayMenu(menu)
	}
	a.Tray.Hide()
}
