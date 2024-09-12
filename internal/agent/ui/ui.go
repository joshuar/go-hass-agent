// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:blank-imports
package ui

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type Agent interface {
	GetMQTTPreferences() *preferences.MQTT
	SaveMQTTPreferences(prefs *preferences.MQTT) error
	GetRestAPIURL() string
	Headless() bool
}

type HassClient interface {
	SensorList() []string
	GetSensor(id string) (sensor.Details, error)
	HassVersion(ctx context.Context) string
}

type Notification interface {
	GetTitle() string
	GetMessage() string
}
