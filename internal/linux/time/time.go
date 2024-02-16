// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package time

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type timeSensor struct {
	linux.Sensor
}

func (s *timeSensor) Attributes() any {
	switch s.SensorTypeValue {
	case linux.SensorUptime:
		return struct {
			NativeUnit string `json:"native_unit_of_measurement"`
			DataSource string `json:"Data Source"`
		}{
			NativeUnit: s.UnitsString,
			DataSource: linux.DataSrcProcfs,
		}
	default:
		return struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: linux.DataSrcProcfs,
		}
	}
}

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 2)
	updateTimes := func(_ time.Duration) {
		sensorCh <- &timeSensor{
			linux.Sensor{
				SensorTypeValue:  linux.SensorUptime,
				Value:            getUptime(ctx),
				IsDiagnostic:     true,
				UnitsString:      "h",
				IconString:       "mdi:restart",
				DeviceClassValue: types.DeviceClassDuration,
				StateClassValue:  types.StateClassMeasurement,
			},
		}
		sensorCh <- &timeSensor{
			linux.Sensor{
				SensorTypeValue:  linux.SensorBoottime,
				Value:            getBoottime(ctx),
				IsDiagnostic:     true,
				IconString:       "mdi:restart",
				DeviceClassValue: types.DeviceClassTimestamp,
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
