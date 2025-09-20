// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"fmt"

	"github.com/joshuar/go-hass-agent/config"
)

// ConfigPrefix is the path prefix in the config for device settings.
const ConfigPrefix = "device"

// Config contains the values that define the device that will be registered with Home Assistant.
type Config struct {
	ID   string `toml:"id"`
	Name string `toml:"name"`
}

// NewConfig returns a new device config.
func NewConfig() error {
	id, err := NewDeviceID()
	if err != nil {
		return fmt.Errorf("unable to generate new device id: %w", err)
	}
	name, err := GetHostname()
	if err != nil {
		return fmt.Errorf("unable to generate device hostname: %w", err)
	}

	err = config.Save(ConfigPrefix, &Config{ID: id, Name: name})
	if err != nil {
		return fmt.Errorf("unable to save device config: %w", err)
	}

	return nil
}

// GetConfig returns the device config.
func GetConfig() (*Config, error) {
	cfg := &Config{}
	if err := config.Load(ConfigPrefix, cfg); err != nil {
		return nil, fmt.Errorf("unable to load agent config: %w", err)
	}
	return cfg, nil
}
