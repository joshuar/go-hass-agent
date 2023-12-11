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
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
)

type timeSensor struct {
	linuxSensor
}

func (s *timeSensor) Attributes() any {
	switch s.sensorType {
	case uptime:
		return struct {
			NativeUnit string `json:"native_unit_of_measurement"`
			DataSource string `json:"Data Source"`
		}{
			NativeUnit: s.units,
			DataSource: srcProcfs,
		}
	default:
		return struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: srcProcfs,
		}
	}
}

func TimeUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 2)
	updateTimes := func(_ time.Duration) {
		sensorCh <- &timeSensor{
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
		sensorCh <- &timeSensor{
			linuxSensor{
				sensorType:  boottime,
				value:       getBoottime(ctx),
				diagnostic:  true,
				icon:        "mdi:restart",
				deviceClass: sensor.Timestamp,
			},
		}
	}

	go helpers.PollSensors(ctx, updateTimes, time.Minute*15, time.Minute)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped time sensors.")
	}()
	return sensorCh
}

func getUptime(ctx context.Context) any {
	u, err := host.UptimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve uptime.")
		return sensor.StateUnknown
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
		return sensor.StateUnknown
	}
	return time.Unix(int64(u), 0).Format(time.RFC3339)
}
