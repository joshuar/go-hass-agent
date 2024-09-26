// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package apps

import (
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	runningAppsIcon  = "mdi:apps"
	runningAppsUnits = "apps"
	runningAppsName  = "Running Apps"
	runningAppsID    = "running_apps"
)

func newRunningAppsSensor(runningApps []string) sensor.Entity {
	return sensor.Entity{
		Name:       runningAppsName,
		Units:      runningAppsUnits,
		StateClass: types.StateClassMeasurement,
		EntityState: &sensor.EntityState{
			ID:    runningAppsID,
			Icon:  runningAppsIcon,
			State: len(runningApps),
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
				"apps":        runningApps,
			},
		},
	}
}
