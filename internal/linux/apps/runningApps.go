// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	runningAppsIcon  = "mdi:apps"
	runningAppsUnits = "apps"
	runningAppsName  = "Running Apps"
)

type runningAppsSensor struct {
	apps []string
	linux.Sensor
}

func (r *runningAppsSensor) Attributes() map[string]any {
	attributes := r.Sensor.Attributes()
	attributes["running_apps"] = r.apps

	return attributes
}

func newRunningAppsSensor() *runningAppsSensor {
	return &runningAppsSensor{
		Sensor: linux.Sensor{
			DisplayName:     runningAppsName,
			IconString:      runningAppsIcon,
			UnitsString:     runningAppsUnits,
			StateClassValue: types.StateClassMeasurement,
		},
	}
}
