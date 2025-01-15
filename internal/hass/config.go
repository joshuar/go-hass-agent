// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package hass

import (
	"context"
	"errors"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

var (
	ErrInvalidEntityConfig = errors.New("entity has invalid config")
	ErrInvalidConfig       = errors.New("invalid config")
)

// Config represents the Home Assistant config. It is the data structure
// returned by the "get_config" API request.
type Config struct {
	Entities              map[string]map[string]any `json:"entities"`
	UnitSystem            units                     `json:"unit_system"`
	ConfigDir             string                    `json:"config_dir"`
	LocationName          string                    `json:"location_name"`
	TimeZone              string                    `json:"time_zone"`
	Version               string                    `json:"version"`
	Components            []string                  `json:"components"`
	WhitelistExternalDirs []string                  `json:"whitelist_external_dirs"`
	Elevation             int                       `json:"elevation"`
	Latitude              float64                   `json:"latitude"`
	Longitude             float64                   `json:"longitude"`
}

type units struct {
	Length      string `json:"length"`
	Mass        string `json:"mass"`
	Temperature string `json:"temperature"`
	Volume      string `json:"volume"`
}

// IsEntityDisabled will check whether the given entity is disabled according to
// Home Assistant.
func (c *Config) IsEntityDisabled(entityID string) (bool, error) {
	if c.Entities == nil {
		return false, ErrInvalidConfig
	}

	if v, ok := c.Entities[entityID]["disabled"]; ok {
		disabledState, ok := v.(bool)
		if !ok {
			return false, nil
		}

		return disabledState, nil
	}

	return false, nil
}

type configRequest struct{}

func (c *configRequest) RequestBody() any {
	return struct {
		Type string `json:"type"`
	}{
		Type: "get_config",
	}
}

func (c *configRequest) Retry() bool {
	return false
}

// Version retrieves the version of Home Assistant.
func Version(ctx context.Context) string {
	config, err := api.Send[Config](ctx, preferences.RestAPIURL(), &configRequest{})
	if err != nil {
		logging.FromContext(ctx).
			Debug("Could not fetch Home Assistant config.",
				slog.Any("error", err))

		return "Unknown"
	}

	return config.Version
}
