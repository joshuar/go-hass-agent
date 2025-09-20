// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mqtt

import (
	"fmt"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/device"
)

const (
	mqttConfigPrefix = "mqtt"
)

type Config struct {
	MQTTServer      string `toml:"server,omitempty" validate:"required,uri" kong:"required,help='MQTT server URI. Required.',placeholder='scheme://some.host:port'"` //nolint:lll
	MQTTUser        string `toml:"user,omitempty" validate:"omitempty" kong:"optional,help='MQTT username.'"`
	MQTTPassword    string `toml:"password,omitempty" validate:"omitempty" kong:"optional,help='MQTT password.'"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" validate:"required,ascii" kong:"optional,help='MQTT topic prefix.'"`
	MQTTEnabled     bool   `toml:"enabled" validate:"boolean" kong:"-"`
}

func (c *Config) Server() string {
	return c.MQTTServer
}

func (c *Config) User() string {
	return c.MQTTUser
}

func (c *Config) Password() string {
	return c.MQTTPassword
}

func (c *Config) TopicPrefix() string {
	return c.MQTTTopicPrefix
}

// MQTTOrigin defines Go Hass Agent as the origin for MQTT functionality.
func MQTTOrigin() *mqtthass.Origin {
	return &mqtthass.Origin{
		Name:    config.AppName,
		Version: config.AppVersion,
		URL:     config.AppURL,
	}
}

// MQTTDevice will return a device that is needed for MQTT functionality.
func MQTTDevice() (*mqtthass.Device, error) {
	// Retrieve the hardware model and manufacturer.
	model, manufacturer, _ := device.GetHWProductInfo()

	id, err := config.Get[string]("device.id")
	if err != nil {
		return nil, fmt.Errorf("unable to load device config: %w", err)
	}

	name, err := config.Get[string]("device.name")
	if err != nil {
		return nil, fmt.Errorf("unable to load device config: %w", err)
	}

	return &mqtthass.Device{
		Name:         name,
		URL:          config.AppURL,
		SWVersion:    config.AppVersion,
		Manufacturer: manufacturer,
		Model:        model,
		Identifiers:  []string{name, id},
	}, nil
}
