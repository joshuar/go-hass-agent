// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type activeAppSensor struct {
	linux.Sensor
}

func newActiveAppSensor() *activeAppSensor {
	newSensor := &activeAppSensor{}
	newSensor.SensorSrc = linux.DataSrcDbus
	newSensor.SensorTypeValue = linux.SensorAppActive
	newSensor.IconString = "mdi:application"

	return newSensor
}
