// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package hass

import (
	"errors"
)

var (
	ErrInvalidEntityConfig = errors.New("entity has invalid config")
	ErrInvalidConfig       = errors.New("invalid config")
)

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

func (c *Config) IsEntityDisabled(entity string) (bool, error) {
	if c.Entities == nil {
		return false, ErrInvalidConfig
	}

	if v, ok := c.Entities[entity]["disabled"]; ok {
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
