// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type activeAppSensor struct {
	linux.Sensor
}

func (a *activeAppSensor) app() string {
	if app, ok := a.State().(string); ok {
		return app
	}
	return ""
}

func (a *activeAppSensor) update(l map[string]dbus.Variant) sensor.Details {
	for app, v := range l {
		if appState, ok := v.Value().(uint32); ok {
			if appState == 2 && a.app() != app {
				a.Value = app
				return a
			}
		}
	}
	return nil
}

func newActiveAppSensor() *activeAppSensor {
	s := &activeAppSensor{}
	s.SensorSrc = linux.DataSrcDbus
	s.SensorTypeValue = linux.SensorAppActive
	s.IconString = "mdi:application"
	return s
}
