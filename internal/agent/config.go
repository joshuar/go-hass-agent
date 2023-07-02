// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"errors"
	"fmt"
	"net/url"
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

func (agent *Agent) IsRegistered() bool {
	return agent.app.Preferences().BoolWithFallback("Registered", false)
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

// agentConfig implements config.Config

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
	var value interface{}
	var err error

	value, _ = c.Get("apiURL")
	err = c.validator.Var(value, "required,url")
	if err != nil {
		log.Debug().Caller().Err(err).Msgf("Validation failure for %s.", "apiURL")
		return errors.New("config validation failed")
	}

	value, _ = c.Get("websocketURL")
	err = c.validator.Var(value, "required,url")
	if err != nil {
		log.Debug().Caller().Err(err).Msgf("Validation failure for %s.", "websocketURL")
		return errors.New("config validation failed")
	}

	value, _ = c.Get("token")
	err = c.validator.Var(value, "required,ascii")
	if err != nil {
		log.Debug().Caller().Err(err).Msgf("Validation failure for %s.", "token")
		return errors.New("config validation failed")
	}

	value, _ = c.Get("webhookID")
	err = c.validator.Var(value, "required,ascii")
	if err != nil {
		log.Debug().Caller().Err(err).Msgf("Validation failure for %s.", "webhookID")
		return errors.New("config validation failed")
	}

	return nil
}

func (c *agentConfig) Refresh() error {
	log.Debug().Caller().
		Msg("Agent config does not support refresh.")
	return nil
}

func (c *agentConfig) generateWebsocketURL() string {
	// TODO: look into websocket http upgrade method
	var scheme string
	host := c.prefs.String("Host")
	u, err := url.Parse(host)
	if err != nil {
		log.Debug().Err(err).Msg("Could not parse host into URL.")
	}
	switch u.Scheme {
	case "https":
		scheme = "wss://"
	case "http":
		fallthrough
	default:
		scheme = "ws://"
	}
	return scheme + u.Host + websocketPath
}

func (c *agentConfig) generateAPIURL() string {
	cloudhookURL := c.prefs.String("CloudhookURL")
	remoteUIURL := c.prefs.String("RemoteUIURL")
	webhookID := c.prefs.String("WebhookID")
	host := c.prefs.String("Host")
	switch {
	case cloudhookURL != "":
		return cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		return remoteUIURL + webHookPath + webhookID
	case webhookID != "" && host != "":
		return host + webHookPath + webhookID
	default:
		return ""
	}
}
