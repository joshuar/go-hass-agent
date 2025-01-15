// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"errors"
	"fmt"
)

const (
	hassPrefPrefix       = "hass"
	prefHassSecret       = hassPrefPrefix + ".secret"
	prefHassAPIURL       = hassPrefPrefix + ".apiurl"
	prefHassWebsocketURL = hassPrefPrefix + ".websocketurl"
	prefHassWebhookID    = hassPrefPrefix + ".webhook_id"
	prefHassCloudhookURL = hassPrefPrefix + ".cloudhook_url"
	prefHassRemoteUIURL  = hassPrefPrefix + ".remote_ui_url"
	regPrefPrefix        = "registration"
	prefHassRegToken     = regPrefPrefix + ".token"
	prefHassRegServer    = regPrefPrefix + ".server"
)

// Hass contains preferences related to connectivity to Home Assistant.
type Hass struct {
	Secret       string `toml:"secret,omitempty" validate:"omitempty,ascii"`
	WebhookID    string `toml:"webhook_id" validate:"required,ascii"`
	RestAPIURL   string `toml:"apiurl" validate:"required,http_url"`
	WebsocketURL string `toml:"websocketurl" validate:"required,uri"`
}

var (
	ErrSaveHassPreferences = errors.New("error saving hass preferences")
	ErrSetHassPreference   = errors.New("could not set hass preference")
)

// SetHassSecret sets the secret value in the preferences store.
func SetHassSecret(secret string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefHassSecret, secret); err != nil {
			return errors.Join(ErrSetHassPreference, err)
		}

		return nil
	}
}

// SetRestAPIURL will generate an appropriate rest API URL with the given hass
// and registration details and save in the preferences store.
func SetRestAPIURL(url string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefHassAPIURL, url); err != nil {
			return errors.Join(ErrSetHassPreference, err)
		}

		return nil
	}
}

// SetWebsocketURL will generate an appropriate websocket API URL from the given
// server.
func SetWebsocketURL(url string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefHassWebsocketURL, url); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}

		return nil
	}
}

// SetWebhookID sets the webhook ID in the preferences.
func SetWebhookID(id string) SetPreference {
	return func() error {
		if id == "" {
			return nil
		}

		if err := prefsSrc.Set(prefHassWebhookID, id); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}

		return nil
	}
}

func SetServer(server string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefHassRegServer, server); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}

		return nil
	}
}

func SetToken(token string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefHassRegToken, token); err != nil {
			return fmt.Errorf("%w: %w", ErrSetHassPreference, err)
		}

		return nil
	}
}

// RestAPIURL retrieves the configured Home Assistant Rest API URL from the
// preferences.
func RestAPIURL() string {
	return prefsSrc.String(prefHassAPIURL)
}

// RestAPIURL retrieves the configured Home Assistant websocket API URL from the
// preferences.
func WebsocketURL() string {
	return prefsSrc.String(prefHassWebsocketURL)
}

// WebhookID retrieves the Go Hass Agent Webhook ID from the
// preferences.
func WebhookID() string {
	return prefsSrc.String(prefHassWebhookID)
}

// Token retrieves the Go Hass Agent long-lived access token from the
// preferences.
func Token() string {
	return prefsSrc.String(prefHassRegToken)
}
