// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	activeAppsIcon = "mdi:application"
	activeAppsName = "Active App"
	activeAppsID   = "active_app"
)

type activeAppSensor struct {
	linux.Sensor
}

func newActiveAppSensor() *activeAppSensor {
	return &activeAppSensor{
		Sensor: linux.Sensor{
			DisplayName: activeAppsName,
			UniqueID:    activeAppsID,
			DataSource:  linux.DataSrcDbus,
			IconString:  activeAppsIcon,
		},
	}
}
