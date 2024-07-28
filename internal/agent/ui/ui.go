// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:blank-imports
package ui

import (
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

type Agent interface {
	Stop()
}

type SensorTracker interface {
	SensorList() []string
	Get(key string) (sensor.Details, error)
}

type Notification interface {
	GetTitle() string
	GetMessage() string
}
