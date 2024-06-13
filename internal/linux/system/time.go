// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	uptimeInterval = 15 * time.Minute
	uptimeJitter   = time.Minute
)

type timeSensor struct {
	linux.Sensor
}

//nolint:exhaustive
func (s *timeSensor) Attributes() any {
	switch s.SensorTypeValue {
	case linux.SensorUptime:
		return struct {
			NativeUnit string `json:"native_unit_of_measurement"`
			DataSource string `json:"data_source"`
		}{
			NativeUnit: s.UnitsString,
			DataSource: linux.DataSrcProcfs,
		}
	default:
		return struct {
			DataSource string `json:"data_source"`
		}{
			DataSource: linux.DataSrcProcfs,
		}
	}
}

type timeWorker struct{}

func (w *timeWorker) Interval() time.Duration { return uptimeInterval }

func (w *timeWorker) Jitter() time.Duration { return uptimeJitter }

//nolint:exhaustruct
func (w *timeWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	return []sensor.Details{
			&timeSensor{
				linux.Sensor{
					SensorTypeValue:  linux.SensorUptime,
					Value:            getUptime(ctx),
					IsDiagnostic:     true,
					UnitsString:      "h",
					IconString:       "mdi:restart",
					DeviceClassValue: types.DeviceClassDuration,
					StateClassValue:  types.StateClassMeasurement,
				},
			},
			&timeSensor{
				linux.Sensor{
					SensorTypeValue:  linux.SensorBoottime,
					Value:            getBoottime(ctx),
					IsDiagnostic:     true,
					IconString:       "mdi:restart",
					DeviceClassValue: types.DeviceClassTimestamp,
				},
			},
		},
		nil
}

func NewTimeWorker(_ context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Time Sensors",
			WorkerDesc: "Sensors for uptime and boottime.",
			Value:      &timeWorker{},
		},
		nil
}

func getUptime(ctx context.Context) any {
	value, err := host.UptimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve uptime.")

		return sensor.StateUnknown
	}

	epoch := time.Unix(0, 0)
	uptime := time.Unix(int64(value), 0)

	return uptime.Sub(epoch).Hours()
}

func getBoottime(ctx context.Context) string {
	value, err := host.BootTimeWithContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to retrieve boottime.")

		return sensor.StateUnknown
	}

	return time.Unix(int64(value), 0).Format(time.RFC3339)
}
