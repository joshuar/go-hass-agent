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
	"github.com/rs/zerolog/log"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

func (agent *Agent) AppConfigVersion() string {
	return agent.app.Preferences().String("Version")
}

func (agent *Agent) DeviceDetails() (string, string) {
	return agent.app.Preferences().String("DeviceName"),
		agent.app.Preferences().String("DeviceID")
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
	case "secret":
		return c.prefs.String("Secret"), nil
	default:
		return nil, fmt.Errorf("unknown config property %s", property)
	}
}

func (c *agentConfig) Set(property string, value interface{}) error {
	valueType := reflect.ValueOf(value)
	switch valueType.Kind() {
	case reflect.String:
		c.prefs.SetString(property, value.(string))
		return nil
	case reflect.Bool:
		c.prefs.SetBool(property, value.(bool))
		return nil
	default:
		return fmt.Errorf("could not set property %s with value %v", property, value)
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
