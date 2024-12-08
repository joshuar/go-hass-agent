// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	"github.com/knadh/koanf/v2"
)

// MQTT contains preferences related to MQTT functionality in Go Hass Agent.
type MQTT struct {
	MQTTServer      string `toml:"server,omitempty" validate:"omitempty,uri" kong:"required,help='MQTT server URI. Required.',placeholder='scheme://some.host:port'"` //nolint:lll
	MQTTUser        string `toml:"user,omitempty" validate:"omitempty" kong:"optional,help='MQTT username.'"`
	MQTTPassword    string `toml:"password,omitempty" validate:"omitempty" kong:"optional,help='MQTT password.'"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" validate:"omitempty,ascii" kong:"optional,help='MQTT topic prefix.'"`
	MQTTEnabled     bool   `toml:"enabled" validate:"boolean" kong:"-"`
}

var ErrSetMQTTPreference = errors.New("could not set MQTT preference")

// SetMQTTPreferences will set Go Hass Agent's MQTT preferences to the given
// values.
func SetMQTTPreferences(prefs *MQTT) error {
	if err := prefsSrc.Set("mqtt.server", prefs.MQTTServer); err != nil {
		return fmt.Errorf("%w: %w", ErrSetMQTTPreference, err)
	}

	if err := prefsSrc.Set("mqtt.user", prefs.MQTTUser); err != nil {
		return fmt.Errorf("%w: %w", ErrSetMQTTPreference, err)
	}

	if err := prefsSrc.Set("mqtt.password", prefs.MQTTPassword); err != nil {
		return fmt.Errorf("%w: %w", ErrSetMQTTPreference, err)
	}

	if err := prefsSrc.Set("mqtt.topic_prefix", prefs.MQTTTopicPrefix); err != nil {
		return fmt.Errorf("%w: %w", ErrSetMQTTPreference, err)
	}

	if err := prefsSrc.Set("mqtt.enabled", prefs.MQTTEnabled); err != nil {
		return fmt.Errorf("%w: %w", ErrSetMQTTPreference, err)
	}

	return nil
}

// GetMQTTPreferences retrieves the current MQTT preferences from file.
func GetMQTTPreferences() (*MQTT, error) {
	var mqttPrefs MQTT
	// Unmarshal config, overwriting defaults.
	if err := prefsSrc.UnmarshalWithConf("mqtt", &mqttPrefs, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}

	return &mqttPrefs, nil
}

// MQTTEnabled will return whether Go Hass Agent will use MQTT.
func MQTTEnabled() bool {
	return prefsSrc.Bool("mqtt.enabled")
}

// Server returns the broker URI from the preferences.
func (p *MQTT) Server() string {
	return p.MQTTServer
}

// User returns any username required for connecting to the broker from the
// preferences.
func (p *MQTT) User() string {
	return p.MQTTUser
}

// Password returns any password required for connecting to the broker from the
// preferences.
func (p *MQTT) Password() string {
	return p.MQTTPassword
}

// TopicPrefix returns the prefix for topics on MQTT.
func (p *MQTT) TopicPrefix() string {
	if p.MQTTTopicPrefix == "" {
		return MQTTTopicPrefix
	}

	return p.MQTTTopicPrefix
}

// MQTTOrigin defines Go Hass Agent as the origin for MQTT functionality.
func MQTTOrigin() *mqtthass.Origin {
	return &mqtthass.Origin{
		Name:    AppName,
		Version: AppVersion,
		URL:     AppURL,
	}
}

// IsMQTTEnabled is a conveinience function to determine whether MQTT
// functionality has been enabled in the agent.
func (p *preferences) IsMQTTEnabled() bool {
	if p.MQTT != nil {
		return p.MQTT.MQTTEnabled
	}

	return false
}
