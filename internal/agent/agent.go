package agent

import (
	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	log "github.com/sirupsen/logrus"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

type Agent struct {
	App    fyne.App
	Tray   fyne.Window
	config config.AppConfig

	Name, Version string
}

func NewAgent() *Agent {
	a := NewUI()
	w := startTrayIcon(a, Name)
	w.SetMaster()
	w.Hide()
	return &Agent{
		App:     a,
		Tray:    w,
		Name:    Name,
		Version: Version,
	}
}

func (a *Agent) LoadConfig() {
	go func() {
		for {
			CloudhookURL := a.App.Preferences().String("CloudhookURL")
			RemoteUIURL := a.App.Preferences().String("RemoteUIURL")
			WebhookID := a.App.Preferences().String("WebhookID")
			InstanceURL := a.App.Preferences().String("InstanceURL")
			switch {
			case CloudhookURL != "":
				a.config.RestAPIURL = CloudhookURL
				log.Debugf("Using CloudhookURL %s for REST API access", a.config.RestAPIURL)
				return
			case RemoteUIURL != "" && WebhookID != "":
				a.config.RestAPIURL = RemoteUIURL + "/api/webhook/" + WebhookID
				log.Debugf("Using RemoteUIURL %s for REST API access", a.config.RestAPIURL)
				return
			case WebhookID != "" && InstanceURL != "":
				a.config.RestAPIURL = InstanceURL + "/api/webhook/" + WebhookID
				log.Debugf("Using InstanceURL %s for REST API access", a.config.RestAPIURL)
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
	}()
}

func (a *Agent) GetConfigVersion() string {
	return a.App.Preferences().String("Version")
}
