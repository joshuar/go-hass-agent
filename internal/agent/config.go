package agent

import (
	"github.com/jeandeaual/go-locale"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

func (agent *Agent) loadConfig() *config.AppConfig {
	CloudhookURL := agent.App.Preferences().String("CloudhookURL")
	RemoteUIURL := agent.App.Preferences().String("RemoteUIURL")
	Host := agent.App.Preferences().String("Host")
	UseTLS := agent.App.Preferences().Bool("UseTLS")

	appConfig := &config.AppConfig{}
	appConfig.Secret = agent.App.Preferences().String("Secret")
	appConfig.Token = agent.App.Preferences().String("Token")
	appConfig.WebhookID = agent.App.Preferences().String("WebhookID")

	if UseTLS {
		appConfig.WebSocketURL = "wss://" + Host + websocketPath
	} else {
		appConfig.WebSocketURL = "ws://" + Host + websocketPath
	}

	switch {
	case CloudhookURL != "":
		appConfig.APIURL = CloudhookURL
		log.Debug().Caller().
			Msgf("Using CloudhookURL %s for Home Assistant access", appConfig.APIURL)
	case RemoteUIURL != "" && appConfig.WebhookID != "":
		appConfig.APIURL = RemoteUIURL + webHookPath + appConfig.WebhookID
		log.Debug().Caller().
			Msgf("Using RemoteUIURL %s for Home Assistant access", appConfig.APIURL)
	case appConfig.WebhookID != "" && Host != "":
		if UseTLS {
			appConfig.APIURL = "https://" + Host + webHookPath + appConfig.WebhookID
		} else {
			appConfig.APIURL = "http://" + Host + webHookPath + appConfig.WebhookID
		}
		log.Debug().Caller().
			Msgf("Using URL %s for Home Assistant access", appConfig.APIURL)
	}
	return appConfig
}

func (agent *Agent) GetConfigVersion() string {
	return agent.App.Preferences().String("Version")
}

func (agent *Agent) GetDeviceDetails() (string, string) {
	return agent.App.Preferences().String("DeviceName"),
		agent.App.Preferences().String("DeviceID")
}

func newMsgPrinter() *message.Printer {
	userLocales, err := locale.GetLocales()
	if err != nil {
		log.Warn().Msg("Could not find a suitable locale. Defaulting to English.")
		return message.NewPrinter(message.MatchLanguage(language.English.String()))
	}
	log.Debug().Caller().Msgf("Setting language to %v.", userLocales)
	return message.NewPrinter(message.MatchLanguage(userLocales...))
}
