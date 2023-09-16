// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ui

import (
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

type Agent interface {
	AppVersion() string
	AppName() string
	AppID() string
	Stop()
	GetConfig(string, interface{}) error
	SetConfig(string, interface{}) error
	SensorList() []string
	SensorValue(string) (tracker.Sensor, error)
}
