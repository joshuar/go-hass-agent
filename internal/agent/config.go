// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"fmt"
	"reflect"

	"fyne.io/fyne/v2"
	"github.com/go-playground/validator/v10"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"
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

type agentConfig struct {
	prefs     fyne.Preferences
	validator *validator.Validate
}

func (agent *Agent) LoadConfig() *agentConfig {
	return &agentConfig{
		prefs:     agent.app.Preferences(),
		validator: validator.New(),
	}
}

func (c *agentConfig) Get(property string) (interface{}, error) {
	switch property {
	case "version":
		return c.prefs.StringWithFallback("Version", Version), nil
	case "websocketURL":
		return c.generateWebsocketURL(), nil
	case "apiURL":
		return c.generateAPIURL(), nil
	case "token":
		return c.prefs.String("Token"), nil
	case "webhookID":
		return c.prefs.String("WebhookID"), nil
	default:
		return nil, fmt.Errorf("unknown config property %s", property)
	}
}

func (c *agentConfig) Set(property string, value interface{}) {
	valueType := reflect.ValueOf(value)
	switch valueType.Kind() {
	case reflect.String:
		c.prefs.SetString(property, value.(string))
	case reflect.Bool:
		c.prefs.SetBool(property, value.(bool))
	}
}

func (c *agentConfig) Validate() error {
	checkErr := func(err error) {
		if err != nil {
			log.Debug().Caller().Err(err).Msg("Validation error.")
			return
		}
	}
	var value interface{}
	var err error

	value, err = c.Get("apiURL")
	checkErr(err)
	err = c.validator.Var(value, "required,url")
	checkErr(err)

	value, err = c.Get("websocketURL")
	checkErr(err)
	err = c.validator.Var(value, "required,url")
	checkErr(err)

	value, err = c.Get("token")
	checkErr(err)
	err = c.validator.Var(value, "required,ascii")
	checkErr(err)

	value, err = c.Get("webhookID")
	checkErr(err)
	err = c.validator.Var(value, "required,ascii")
	checkErr(err)

	return nil
}

func (c *agentConfig) generateWebsocketURL() string {
	var scheme string
	if c.prefs.BoolWithFallback("UseTLS", false) {
		scheme = "wss://"
	} else {
		scheme = "ws://"
	}
	return scheme + c.prefs.StringWithFallback("Host", "localhost") + websocketPath
}

func (c *agentConfig) generateAPIURL() string {
	cloudhookURL := c.prefs.String("CloudhookURL")
	remoteUIURL := c.prefs.String("RemoteUIURL")
	webhookID := c.prefs.String("WebhookID")
	host := c.prefs.String("Host")
	useTLS := c.prefs.Bool("UseTLS")
	switch {
	case cloudhookURL != "":
		return cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		return remoteUIURL + webHookPath + webhookID
	case webhookID != "" && host != "":
		var scheme string
		if useTLS {
			scheme = "https://"
		} else {
			scheme = "http://"
		}
		return scheme + host + webHookPath + webhookID
	default:
		return ""
	}
}
