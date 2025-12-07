// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"errors"
	"sync"

	"github.com/joshuar/go-hass-agent/hass/api"
)

const (
	ConfigPrefix       = "hass"
	ConfigAPIURL       = "apiurl"
	ConfigWebsocketURL = "websocketurl"
	ConfigWebhookID    = "webhook_id"
	ConfigSecret       = "secret"
)

type Config struct {
	mu           sync.Mutex          `toml:"-"`
	APIURL       string              `toml:"apiurl"       validate:"required"`
	Secret       string              `toml:"secret"`
	WebHookID    string              `toml:"webhook_id"   validate:"required"`
	WebsocketURL string              `toml:"websocketurl" validate:"required"`
	remote       *api.ConfigResponse `toml:"-"`
}

var (
	ErrInvalidEntityConfig = errors.New("entity has invalid config")
	ErrInvalidConfig       = errors.New("invalid config")
)

func (c *Config) Update(newConfig *api.ConfigResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.remote = newConfig
}

// GetVersion returns the version of Home Assistant from the config response, or
// "Unknown" if it was not set.
func (c *Config) GetVersion() string {
	return c.remote.Version
}

// IsEntityDisabled returns whether the entity with the given ID has been
// disabled in Home Assistant. If the disabled status could not be determined,
// it will return a non-nil error.
func (c *Config) IsEntityDisabled(id string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If there is no entities list, assume not disabled.
	if c.remote.Entities.IsNull() {
		return false, nil
	}

	entities, err := c.remote.Entities.Get()
	if err != nil {
		return false, errors.Join(ErrInvalidConfig, err)
	}

	if v, ok := entities[id]["disabled"]; ok {
		disabledState, ok := v.(bool)
		if !ok {
			return false, nil
		}
		return disabledState, nil
	}

	return false, nil
}
