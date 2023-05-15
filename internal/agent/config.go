// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"reflect"

	"github.com/joshuar/go-hass-agent/internal/config"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

func (agent *Agent) loadAppConfig() *config.AppConfig {
	var CloudhookURL, RemoteUIURL, Host string
	var UseTLS bool

	agent.Pref("CloudhookURL", &CloudhookURL)
	agent.Pref("RemoteUIURL", &RemoteUIURL)
	agent.Pref("Host", &Host)
	agent.Pref("UseTLS", &UseTLS)

	appConfig := &config.AppConfig{}
	agent.Pref("Token", &appConfig.Token)
	agent.Pref("Secret", &appConfig.Secret)
	agent.Pref("WebhookID", &appConfig.WebhookID)

	var scheme string
	if UseTLS {
		scheme = "wss://"
	} else {
		scheme = "ws://"
	}
	appConfig.WebSocketURL = scheme + Host + websocketPath

	switch {
	case CloudhookURL != "":
		appConfig.APIURL = CloudhookURL
	case RemoteUIURL != "" && appConfig.WebhookID != "":
		appConfig.APIURL = RemoteUIURL + webHookPath + appConfig.WebhookID
	case appConfig.WebhookID != "" && Host != "":
		var scheme string
		if UseTLS {
			scheme = "https://"
		} else {
			scheme = "http://"
		}
		appConfig.APIURL = scheme + Host + webHookPath + appConfig.WebhookID
	}

	return appConfig
}

func (agent *Agent) AppConfigVersion() string {
	return agent.app.Preferences().String("Version")
}

func (agent *Agent) DeviceDetails() (string, string) {
	return agent.app.Preferences().String("DeviceName"),
		agent.app.Preferences().String("DeviceID")
}

func (agent *Agent) Pref(pref string, value interface{}) {
	valueType := reflect.ValueOf(value).Elem()
	switch valueType.Kind() {
	case reflect.String:
		newValue := value.(*string)
		*newValue = agent.app.Preferences().String(pref)
		value = newValue
	case reflect.Bool:
		newValue := value.(*bool)
		*newValue = agent.app.Preferences().Bool(pref)
		value = newValue
	}
}

func (agent *Agent) SetPref(pref string, value interface{}) {
	valueType := reflect.ValueOf(value)
	switch valueType.Kind() {
	case reflect.String:
		agent.app.Preferences().SetString(pref, value.(string))
	case reflect.Bool:
		agent.app.Preferences().SetBool(pref, value.(bool))
	}
}
