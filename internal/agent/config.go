// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

func (agent *Agent) loadAppConfig() *config.AppConfig {
	CloudhookURL := agent.app.Preferences().String("CloudhookURL")
	RemoteUIURL := agent.app.Preferences().String("RemoteUIURL")
	Host := agent.app.Preferences().String("Host")
	UseTLS := agent.app.Preferences().Bool("UseTLS")

	appConfig := &config.AppConfig{}
	appConfig.Secret = agent.app.Preferences().String("Secret")
	appConfig.Token = agent.app.Preferences().String("Token")
	appConfig.WebhookID = agent.app.Preferences().String("WebhookID")

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

func (agent *Agent) GetAppConfigVersion() string {
	return agent.app.Preferences().String("Version")
}

func (agent *Agent) GetDeviceDetails() (string, string) {
	return agent.app.Preferences().String("DeviceName"),
		agent.app.Preferences().String("DeviceID")
}
