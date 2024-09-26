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
	activeAppsIcon = "mdi:application"
	activeAppsName = "Active App"
	activeAppsID   = "active_app"
)

func newActiveAppSensor(name string) sensor.Entity {
	return sensor.Entity{
		Name: activeAppsName,
		EntityState: &sensor.EntityState{
			ID:         activeAppsID,
			Icon:       activeAppsIcon,
			State:      name,
			EntityType: types.Sensor,
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
			},
		},
	}
}
