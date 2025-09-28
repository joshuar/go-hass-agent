// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mqtt

import (
	"fmt"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/device"
	"github.com/joshuar/go-hass-agent/validation"
)

const (
	// ConfigPrefix is the prefix in the configuration file for MQTT preferences.
	ConfigPrefix = "mqtt"
)

// Config represents MQTT preferences.
type Config struct {
	MQTTServer      string `toml:"server,omitempty" form:"mqtt_server" validate:"required,uri" kong:"required,help='MQTT server URI. Required.',placeholder='scheme://some.host:port'"` //nolint:lll
	MQTTUser        string `toml:"user,omitempty" form:"mqtt_user" validate:"omitempty" kong:"optional,help='MQTT username.'"`
	MQTTPassword    string `toml:"password,omitempty" form:"mqtt_password" validate:"omitempty" kong:"optional,help='MQTT password.'"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" form:"mqtt_topic_prefix" validate:"required,ascii" kong:"optional,help='MQTT topic prefix.'"`
	MQTTEnabled     bool   `toml:"enabled" form:"mqtt_enabled" validate:"boolean" kong:"-"`
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

// Valid will check the MQTT preferences are valid.
func (c *Config) Valid() (bool, error) {
	err := validation.Validate.Struct(c)
	if err != nil {
		return false, fmt.Errorf("%w: %s", validation.ErrValidation, validation.ParseValidationErrors(err))
	}

	return true, nil
}

// Sanitise will sanitise the values of the MQTT preferences.
func (c *Config) Sanitise() error {
	return nil
}

// Origin defines Go Hass Agent as the origin for MQTT functionality.
func Origin() *mqtthass.Origin {
	return &mqtthass.Origin{
		Name:    config.AppName,
		Version: config.AppVersion,
		URL:     config.AppURL,
	}
}

// Device will return a device that is needed for MQTT functionality.
func Device() (*mqtthass.Device, error) {
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
