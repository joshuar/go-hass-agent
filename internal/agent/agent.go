package agent

import (
	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

type Agent struct {
	App    fyne.App
	config config.AppConfig

	Name, Version string
}

func NewAgent() *Agent {
	return &Agent{
		App:     NewUI(),
		Name:    Name,
		Version: Version,
	}
}

func (a *Agent) LoadConfig() {

	done := make(chan bool, 1)

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
				done <- true
			case RemoteUIURL != "" && WebhookID != "":
				a.config.RestAPIURL = RemoteUIURL + "/api/webhook/" + WebhookID
				log.Debugf("Using RemoteUIURL %s for REST API access", a.config.RestAPIURL)
				done <- true
			case WebhookID != "" && InstanceURL != "":
				log.Debugf("Using InstanceURL %s for REST API access", a.config.RestAPIURL)
				a.config.RestAPIURL = InstanceURL + "/api/webhook/" + WebhookID
				done <- true
			default:
				log.Warn("No suitable existing config found, running registration")
				userRegistrationInfo := a.GetRegistrationInfo()
				appRegistrationInfo := a.RegisterWithHass(userRegistrationInfo)
				a.SaveRegistration(appRegistrationInfo)
			}
		}
	}()
	<-done
}

func (a *Agent) GetConfigVersion() string {
	return a.App.Preferences().String("Version")
}
