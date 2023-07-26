// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
)

type timeSensor struct {
	linuxSensor
}

func (s *timeSensor) Attributes() interface{} {
	switch s.sensorType {
	case uptime:
		return struct {
			NativeUnit string `json:"native_unit_of_measurement"`
			DataSource string `json:"Data Source"`
		}{
			NativeUnit: s.units,
			DataSource: "procfs",
		}
	default:
		return struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: "procfs",
		}
	}
}

func TimeUpdater(ctx context.Context, status chan interface{}) {
	updateTimes := func() {
		status <- &timeSensor{
			linuxSensor{
				sensorType:  uptime,
				value:       getUptime(ctx),
				diagnostic:  true,
				units:       "h",
				icon:        "mdi:restart",
				deviceClass: sensor.Duration,
				stateClass:  sensor.StateMeasurement,
			},
		}

		status <- &timeSensor{
			linuxSensor{
				sensorType:  boottime,
				value:       getBoottime(ctx),
				diagnostic:  true,
				icon:        "mdi:restart",
				deviceClass: sensor.Timestamp,
			},
		}
	}

	helpers.PollSensors(ctx, updateTimes, time.Minute*15, time.Minute)
}

func getUptime(ctx context.Context) interface{} {
	u, err := host.UptimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve uptime.")
		return "Unknown"
	}
	epoch := time.Unix(0, 0)
	uptime := time.Unix(int64(u), 0)
	return uptime.Sub(epoch).Hours()
}

func getBoottime(ctx context.Context) string {
	u, err := host.BootTimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve boottime.")
		return "Unknown"
	}
	return time.Unix(int64(u), 0).Format(time.RFC3339)
}
