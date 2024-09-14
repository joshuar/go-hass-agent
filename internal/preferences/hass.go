// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"fmt"
	"net/url"
)

const (
	WebsocketPath = "/api/websocket"
	WebHookPath   = "/api/webhook/"
)

type Hass struct {
	CloudhookURL string `toml:"cloudhook_url,omitempty" json:"cloudhook_url" validate:"omitempty,http_url"`
	RemoteUIURL  string `toml:"remote_ui_url,omitempty" json:"remote_ui_url" validate:"omitempty,http_url"`
	Secret       string `toml:"secret,omitempty" json:"secret" validate:"omitempty,ascii"`
	WebhookID    string `toml:"webhook_id" json:"webhook_id" validate:"required,ascii"`
	RestAPIURL   string `toml:"apiurl,omitempty" json:"-" validate:"required_without=CloudhookURL RemoteUIURL,http_url"`
	WebsocketURL string `toml:"websocketurl" json:"-" validate:"required,url"`
}

func DefaultHassPreferences() *Hass {
	return &Hass{
		RestAPIURL:   DefaultServer,
		WebsocketURL: DefaultServer,
		WebhookID:    DefaultSecret,
	}
}

func (p *Preferences) SaveHassPreferences(prefs *Hass, options *Registration) error {
	p.Hass = prefs
	p.Registration = options

	if err := p.generateAPIURL(); err != nil {
		return fmt.Errorf("unable generate API URL: %w", err)
	}

	// Generate a websocket URL.
	if err := p.generateWebsocketURL(); err != nil {
		return fmt.Errorf("unable to generated websocket URL: %w", err)
	}

	// Set agent as registered
	p.Registered = true

	return p.Save()
}

func (p *Preferences) generateAPIURL() error {
	switch {
	case p.Hass.CloudhookURL != "" && !p.Registration.IgnoreHassURLs:
		p.Hass.RestAPIURL = p.Hass.CloudhookURL
	case p.Hass.RemoteUIURL != "" && p.Hass.WebhookID != "" && !p.Registration.IgnoreHassURLs:
		p.Hass.RestAPIURL = p.Hass.RemoteUIURL + WebHookPath + p.Hass.WebhookID
	default:
		apiURL, err := url.Parse(p.Registration.Server)
		if err != nil {
			return fmt.Errorf("could not parse registration server: %w", err)
		}

		apiURL = apiURL.JoinPath(WebHookPath, p.Hass.WebhookID)

		p.Hass.RestAPIURL = apiURL.String()
	}

	return nil
}

func (p *Preferences) generateWebsocketURL() error {
	websocketURL, err := url.Parse(p.Registration.Server)
	if err != nil {
		return fmt.Errorf("could not parse registration server: %w", err)
	}

	switch websocketURL.Scheme {
	case "https":
		websocketURL.Scheme = "wss"
	case "http":
		websocketURL.Scheme = "ws"
	case "wss":
	default:
		websocketURL.Scheme = "ws"
	}

	websocketURL = websocketURL.JoinPath(WebsocketPath)

	p.Hass.WebsocketURL = websocketURL.String()

	return nil
}
