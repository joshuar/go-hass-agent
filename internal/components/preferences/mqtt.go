// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/device/info"
)

const (
	mqttPrefPrefix      = "mqtt"
	prefMQTTServer      = mqttPrefPrefix + ".server"
	prefMQTTUser        = mqttPrefPrefix + ".user"
	prefMQTTPass        = mqttPrefPrefix + ".password"
	prefMQTTTopicPrefix = mqttPrefPrefix + ".topic_prefix"
	prefMQTTEnabled     = mqttPrefPrefix + ".enabled"
)

// MQTTPreferences contains preferences related to MQTTPreferences functionality in Go Hass Agent.
//
//nolint:lll
type MQTTPreferences struct {
	MQTTServer      string `toml:"server,omitempty" validate:"required,uri" kong:"required,help='MQTT server URI. Required.',placeholder='scheme://some.host:port'"` //nolint:lll
	MQTTUser        string `toml:"user,omitempty" validate:"omitempty" kong:"optional,help='MQTT username.'"`
	MQTTPassword    string `toml:"password,omitempty" validate:"omitempty" kong:"optional,help='MQTT password.'"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" validate:"required,ascii" kong:"optional,help='MQTT topic prefix.'"`
	MQTTEnabled     bool   `toml:"enabled" validate:"boolean" kong:"-"`
}

var (
	// ErrMQTTPreference indicates a problem with setting an MQTT preference.
	ErrMQTTPreference      = errors.New("MQTT preference error")
	defaultMQTTPreferences = &MQTTPreferences{
		MQTTEnabled:     false,
		MQTTTopicPrefix: defaultMQTTTopicPrefix,
		MQTTServer:      defaultMQTTServer,
	}
)

// SetMQTTEnabled will set the MQTT whether MQTT functionality is enabled.
func SetMQTTEnabled(value bool) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefMQTTEnabled, value); err != nil {
			return errors.Join(ErrMQTTPreference, err)
		}

		return nil
	}
}

// SetMQTTServer will set the MQTT server preference.
func SetMQTTServer(server string) SetPreference {
	return func() error {
		if server == "" {
			server = defaultMQTTServer
		}

		if err := prefsSrc.Set(prefMQTTServer, server); err != nil {
			return errors.Join(ErrMQTTPreference, err)
		}

		return nil
	}
}

// SetMQTTTopicPrefix will set the MQTT server preference.
func SetMQTTTopicPrefix(prefix string) SetPreference {
	return func() error {
		if prefix == "" {
			prefix = defaultMQTTTopicPrefix
		}

		if err := prefsSrc.Set(prefMQTTTopicPrefix, prefix); err != nil {
			return errors.Join(ErrMQTTPreference, err)
		}

		return nil
	}
}

// SetMQTTUser will set the MQTT user preference.
func SetMQTTUser(user string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefMQTTUser, user); err != nil {
			return errors.Join(ErrMQTTPreference, err)
		}

		return nil
	}
}

// SetMQTTPassword will set the MQTT password.
func SetMQTTPassword(password string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefMQTTPass, password); err != nil {
			return errors.Join(ErrMQTTPreference, err)
		}

		return nil
	}
}

// MQTTEnabled will return whether Go Hass Agent will use MQTT.
func MQTTEnabled() bool {
	return prefsSrc.Bool(prefMQTTEnabled)
}

// Server returns the broker URI from the preferences.
func (p *MQTTPreferences) Server() string {
	return p.MQTTServer
}

// User returns any username required for connecting to the broker from the
// preferences.
func (p *MQTTPreferences) User() string {
	return p.MQTTUser
}

// Password returns any password required for connecting to the broker from the
// preferences.
func (p *MQTTPreferences) Password() string {
	return p.MQTTPassword
}

// TopicPrefix returns the prefix for topics on MQTT.
func (p *MQTTPreferences) TopicPrefix() string {
	return p.MQTTTopicPrefix
}

// MQTTOrigin defines Go Hass Agent as the origin for MQTT functionality.
func MQTTOrigin() *mqtthass.Origin {
	return &mqtthass.Origin{
		Name:    AppName,
		Version: AppVersion(),
		URL:     AppURL,
	}
}

// MQTT retrieves the current MQTT preferences from file.
func MQTT() *MQTTPreferences {
	prefs := *defaultMQTTPreferences
	if err := load(mqttPrefPrefix, &prefs); err != nil {
		logger.Debug("Could not retrieve MQTT preferences, defaults will be used.",
			slog.Any("error", err))
	}

	return &prefs
}

// MQTTDevice will return a device that is needed for MQTT functionality.
func MQTTDevice() *mqtthass.Device {
	// Retrieve the hardware model and manufacturer.
	model, manufacturer, _ := info.GetHWProductInfo()

	return &mqtthass.Device{
		Name:         prefsSrc.String(prefDeviceName),
		URL:          AppURL,
		SWVersion:    AppVersion(),
		Manufacturer: manufacturer,
		Model:        model,
		Identifiers:  []string{appID, prefsSrc.String(prefDeviceName), prefsSrc.String(prefDeviceID)},
	}
}
