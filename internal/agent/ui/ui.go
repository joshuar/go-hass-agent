// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:blank-imports
package ui

import (
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type Agent interface {
	GetMQTTPreferences() *preferences.MQTT
	SaveMQTTPreferences(prefs *preferences.MQTT) error
	Stop()
	Headless() bool
}

type SensorTracker interface {
	SensorList() []string
	Get(key string) (sensor.Details, error)
}

type Notification interface {
	GetTitle() string
	GetMessage() string
}
