// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mqtt

import (
	"errors"
	"fmt"
	"net/url"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/device"
	"github.com/joshuar/go-hass-agent/validation"
)

const (
	// ConfigPrefix is the prefix in the configuration file for MQTT preferences.
	ConfigPrefix       = "mqtt"
	DefaultTopicPrefix = "homeassistant"
	defaultMQTTServer  = "tcp://localhost:1883"
)

// Config represents MQTT preferences.
type Config struct {
	MQTTServer      string `toml:"server,omitempty"       form:"mqtt.mqtt_server"       validate:"required_if=MQTTEnabled true,omitempty,uri" kong:"help='MQTT server URI.',placeholder='tcp://some.host:port'"`
	MQTTUser        string `toml:"user,omitempty"         form:"mqtt.mqtt_user"         validate:"omitempty"                                  kong:"optional,help='MQTT username.'"`
	MQTTPassword    string `toml:"password,omitempty"     form:"mqtt.mqtt_password"     validate:"omitempty"                                  kong:"optional,help='MQTT password.'"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" form:"mqtt.mqtt_topic_prefix" validate:"required,ascii"                             kong:"optional,help='MQTT topic prefix.'"`
	MQTTEnabled     bool   `toml:"enabled"                form:"mqtt.mqtt_enabled"      validate:"boolean"                                    kong:"negatable,help='Enable MQTT features.'"`
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
	if err := validation.ValidateStruct(c); err != nil {
		return false, fmt.Errorf("validate config: %w", err)
	}
	return true, nil
}

// Sanitise will sanitise the values of the MQTT preferences.
func (c *Config) Sanitise() error {
	if c == nil {
		return errors.New("no config found")
	}
	server, err := url.Parse(c.MQTTServer)
	if err != nil {
		return fmt.Errorf("could not sanitise server value: %w", err)
	}
	// Set scheme to tcp, a common error is to use http or https.
	if server.Scheme != "tcp" {
		server.Scheme = "tcp"
	}
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
		return nil, fmt.Errorf("unable to get a device id: %w", err)
	}

	name, err := config.Get[string]("device.name")
	if err != nil {
		return nil, fmt.Errorf("unable to get a device name: %w", err)
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
