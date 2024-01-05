// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

type activeAppSensor struct {
	linux.Sensor
}

type activeAppSensorAttributes struct {
	DataSource string `json:"Data Source"`
}

func (a *activeAppSensor) Attributes() any {
	if _, ok := a.Value.(string); ok {
		return &activeAppSensorAttributes{
			DataSource: linux.DataSrcDbus,
		}
	}
	return nil
}

func (a *activeAppSensor) app() string {
	if app, ok := a.State().(string); ok {
		return app
	}
	return ""
}

func (a *activeAppSensor) update(l map[string]dbus.Variant, s chan tracker.Sensor) {
	for app, v := range l {
		if appState, ok := v.Value().(uint32); ok {
			if appState == 2 && a.app() != app {
				a.Value = app
				s <- a
			}
		}
	}
}

func newActiveAppSensor() *activeAppSensor {
	s := &activeAppSensor{}
	s.SensorTypeValue = linux.SensorAppActive
	s.IconString = "mdi:application"
	return s
}
