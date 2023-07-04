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
	"time"

	"fyne.io/fyne/v2"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
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
		return c.prefs.String("WebSocketURL"), nil
	case "apiURL":
		return c.prefs.String("ApiURL"), nil
	case "token":
		return c.prefs.String("Token"), nil
	case "webhookID":
		return c.prefs.String("WebhookID"), nil
	case "secret":
		return c.prefs.String("Secret"), nil
	case "host":
		return c.prefs.String("Host"), nil
	case "useTLS":
		return c.prefs.Bool("UseTLS"), nil
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
	if c.validator.Var(value, "required,url") != nil {
		return errors.New("apiURL does not match either a URL, hostname or hostname:port")
	}

	value, _ = c.Get("websocketURL")
	if c.validator.Var(value, "required,url") != nil {
		return errors.New("websocketURL does not match either a URL, hostname or hostname:port")
	}

	value, _ = c.Get("token")
	if err = c.validator.Var(value, "required,ascii"); err != nil {
		return errors.New("invalid long-lived token format")
	}

	value, _ = c.Get("webhookID")
	if err = c.validator.Var(value, "required,ascii"); err != nil {
		return errors.New("invalid webhookID format")
	}

	return nil
}

func (c *agentConfig) Refresh() error {
	log.Debug().Caller().
		Msg("Agent config does not support refresh.")
	return nil
}

func (c *agentConfig) Upgrade() error {
	configVersion, err := c.Get("version")
	if err != nil {
		return err
	}
	switch {
	// * Upgrade host to include scheme for versions < v.1.4.0
	case semver.Compare(configVersion.(string), "v1.4.0") < 0:
		host, err := c.Get("host")
		if err != nil {
			return err
		}
		useTLS, err := c.Get("useTLS")
		if err != nil {
			return err
		}
		var newHost string
		switch useTLS.(type) {
		case bool:
			if useTLS.(bool) {
				newHost = "https://" + host.(string)
			} else {
				newHost = "http://" + host.(string)
			}
		default:
			newHost = "http://" + host.(string)
		}
		c.Set("Host", newHost)
		fallthrough
	// * Add ApiURL and WebSocketURL config options for versions < v1.4.3
	case semver.Compare(configVersion.(string), "v1.4.3") < 0:
		c.generateAPIURL()
		c.generateWebsocketURL()
	}

	c.Set("Version", Version)

	// ! https://github.com/fyne-io/fyne/issues/3170
	time.Sleep(110 * time.Millisecond)

	return nil
}

func (c *agentConfig) generateWebsocketURL() {
	// TODO: look into websocket http upgrade method
	host := c.prefs.String("Host")
	url, _ := url.Parse(host)
	switch url.Scheme {
	case "https":
		url.Scheme = "wss"
	case "http":
		fallthrough
	default:
		url.Scheme = "ws"
	}
	url = url.JoinPath(websocketPath)

	c.Set("WebSocketURL", url.String())
}

func (c *agentConfig) generateAPIURL() {
	cloudhookURL := c.prefs.String("CloudhookURL")
	remoteUIURL := c.prefs.String("RemoteUIURL")
	webhookID := c.prefs.String("WebhookID")
	host := c.prefs.String("Host")
	var apiURL string
	switch {
	case cloudhookURL != "":
		apiURL = cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		apiURL = remoteUIURL + webHookPath + webhookID
	case webhookID != "" && host != "":
		url, _ := url.Parse(host)
		url = url.JoinPath(webHookPath, webhookID)
		apiURL = url.String()
	default:
		apiURL = ""
	}
	c.Set("ApiURL", apiURL)
}
