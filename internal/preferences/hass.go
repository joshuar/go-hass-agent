// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	WebsocketPath = "/api/websocket"
	WebHookPath   = "/api/webhook/"
)

// Hass contains preferences related to connectivity to Home Assistant.
type Hass struct {
	CloudhookURL string `toml:"cloudhook_url,omitempty" json:"cloudhook_url" validate:"omitempty,http_url"`
	RemoteUIURL  string `toml:"remote_ui_url,omitempty" json:"remote_ui_url" validate:"omitempty,http_url"`
	Secret       string `toml:"secret,omitempty" json:"secret" validate:"omitempty,ascii"`
	WebhookID    string `toml:"webhook_id" json:"webhook_id" validate:"required,ascii"`
	RestAPIURL   string `toml:"apiurl,omitempty" json:"-" validate:"required_without=CloudhookURL RemoteUIURL,http_url"`
	WebsocketURL string `toml:"websocketurl" json:"-" validate:"required,url"`
}

var (
	ErrSaveHassPreferences = errors.New("error saving hass preferences")
	ErrSetHassPreference   = errors.New("could not set hass preference")
)

// SetHassPreferences will set the Hass preferences to the given values.
//
//revive:disable:indent-error-flow
func SetHassPreferences(hassPrefs *Hass, regPrefs *Registration) error {
	if err := prefsSrc.Set("hass.secret", hassPrefs.Secret); err != nil {
		return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
	}

	// Generate an API URL and set preferences as appropriate.
	if apiURL, err := generateAPIURL(hassPrefs, regPrefs); err != nil {
		return fmt.Errorf("%w: %w", ErrSaveHassPreferences, err)
	} else {
		if err := prefsSrc.Set("hass.apiurl", apiURL); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}
	}

	// Generate a websocket URL and set preferences as appropriate.
	if websocketURL, err := generateWebsocketURL(regPrefs); err != nil {
		return fmt.Errorf("%w: %w", ErrSaveHassPreferences, err)
	} else {
		if err := prefsSrc.Set("hass.websocketurl", websocketURL); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}
	}

	// Set the webhookID if present.
	if hassPrefs.WebhookID != "" {
		if err := prefsSrc.Set("hass.webhook_id", hassPrefs.WebhookID); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}
	}

	if err := prefsSrc.Set("registration.server", regPrefs.Server); err != nil {
		return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
	}

	if err := prefsSrc.Set("registration.token", regPrefs.Token); err != nil {
		return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
	}

	return nil
}

// RestAPIURL retrieves the configured Home Assistant Rest API URL from the
// preferences.
func RestAPIURL() string {
	return prefsSrc.String("hass.apiurl")
}

// RestAPIURL retrieves the configured Home Assistant websocket API URL from the
// preferences.
func WebsocketURL() string {
	return prefsSrc.String("hass.websocketurl")
}

// WebhookID retrieves the Go Hass Agent Webhook ID from the
// preferences.
func WebhookID() string {
	return prefsSrc.String("hass.webhook_id")
}

// Token retrieves the Go Hass Agent long-lived access token from the
// preferences.
func Token() string {
	return prefsSrc.String("registration.token")
}

func generateAPIURL(hassPrefs *Hass, regPrefs *Registration) (string, error) {
	switch {
	case hassPrefs.CloudhookURL != "" && regPrefs.IgnoreHassURLs:
		if err := prefsSrc.Set("hass.cloudhook_url", hassPrefs.CloudhookURL); err != nil {
			return "", fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}

		return hassPrefs.CloudhookURL, nil
	case hassPrefs.RemoteUIURL != "" && hassPrefs.WebhookID != "" && !regPrefs.IgnoreHassURLs:
		if err := prefsSrc.Set("hass.remote_ui_url", hassPrefs.CloudhookURL); err != nil {
			return "", fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}

		return hassPrefs.RemoteUIURL + WebHookPath + hassPrefs.WebhookID, nil
	default:
		apiURL, err := url.Parse(regPrefs.Server)
		if err != nil {
			return "", fmt.Errorf("could not parse registration server: %w", err)
		}

		return apiURL.JoinPath(WebHookPath, hassPrefs.WebhookID).String(), nil
	}
}

func generateWebsocketURL(regPrefs *Registration) (string, error) {
	websocketURL, err := url.Parse(regPrefs.Server)
	if err != nil {
		return "", fmt.Errorf("could not parse registration server: %w", err)
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

	return websocketURL.JoinPath(WebsocketPath).String(), nil
}
