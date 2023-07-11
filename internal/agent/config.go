// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"github.com/go-playground/validator/v10"
	"github.com/joshuar/go-hass-agent/internal/settings"
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
	if v, ok := value.(string); ok {
		agent.app.Preferences().SetString(pref, v)
		return
	}
	if v, ok := value.(bool); ok {
		agent.app.Preferences().SetBool(pref, v)
		return
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

func (c *agentConfig) Validate() error {
	var err error

	if c.validator.Var(c.prefs.String("ApiURL"), "required,url") != nil {
		return errors.New("apiURL does not match either a URL, hostname or hostname:port")
	}

	if c.validator.Var(c.prefs.String("WebSocketURL"), "required,url") != nil {
		return errors.New("websocketURL does not match either a URL, hostname or hostname:port")
	}

	if err = c.validator.Var(c.prefs.String("Token"), "required,ascii"); err != nil {
		return errors.New("invalid long-lived token format")
	}

	if err = c.validator.Var(c.prefs.String("WebhookID"), "required,ascii"); err != nil {
		return errors.New("invalid webhookID format")
	}

	return nil
}

func (c *agentConfig) Refresh(ctx context.Context) error {
	log.Debug().Caller().
		Msg("Agent config does not support refresh.")
	return nil
}

func (c *agentConfig) Upgrade() error {
	configVersion := c.prefs.String("Version")

	if configVersion == "" {
		return errors.New("config version is not a valid value")
	}
	switch {
	// * Upgrade host to include scheme for versions < v.1.4.0
	case semver.Compare(configVersion, "v1.4.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.0")
		hostString := c.prefs.String("Host")
		if hostString == "" {
			return errors.New("upgrade < v.1.4.0: invalid host value")
		}
		switch c.prefs.Bool("UseTLS") {
		case true:
			hostString = "https://" + hostString
		case false:
			hostString = "http://" + hostString
		}
		c.prefs.SetString("Host", hostString)
		fallthrough
	// * Add ApiURL and WebSocketURL config options for versions < v1.4.3
	case semver.Compare(configVersion, "v1.4.3") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.3")
		c.generateAPIURL()
		c.generateWebsocketURL()
	}

	c.prefs.SetString("Version", Version)

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
	c.prefs.SetString("WebSocketURL", url.String())
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
	c.prefs.SetString("ApiURL", apiURL)
}

func (c *agentConfig) StoreSettings(ctx context.Context) context.Context {
	s := settings.NewSettings()
	s.SetValue("apiURL", c.prefs.String("ApiURL"))
	s.SetValue("webSocketURL", c.prefs.String("WebSocketURL"))
	s.SetValue("secret", c.prefs.String("Secret"))
	s.SetValue("webhookID", c.prefs.String("WebHookID"))
	s.SetValue("token", c.prefs.String("Token"))
	return settings.StoreInContext(ctx, s)
}
