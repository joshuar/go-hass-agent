// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type runningAppsSensor struct {
	apps []string
	linux.Sensor
}

func (r *runningAppsSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["running_apps"] = r.apps
	attributes["data_source"] = linux.DataSrcDbus

	return attributes
}

func newRunningAppsSensor() *runningAppsSensor {
	newSensor := &runningAppsSensor{}
	newSensor.SensorTypeValue = linux.SensorAppRunning
	newSensor.IconString = "mdi:apps"
	newSensor.UnitsString = "apps"
	newSensor.StateClassValue = types.StateClassMeasurement

	return newSensor
}
