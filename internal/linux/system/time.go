// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"log/slog"
	"time"

	"github.com/shirou/gopsutil/v3/host"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	uptimeInterval = 15 * time.Minute
	uptimeJitter   = time.Minute

	timeWorkerID = "time_sensors"
)

type timeSensor struct {
	linux.Sensor
}

type timeWorker struct {
	logger *slog.Logger
}

func (w *timeWorker) Interval() time.Duration { return uptimeInterval }

func (w *timeWorker) Jitter() time.Duration { return uptimeJitter }

func (w *timeWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	return []sensor.Details{
			&timeSensor{
				linux.Sensor{
					SensorTypeValue:  linux.SensorUptime,
					Value:            w.getUptime(ctx),
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
					Value:            w.getBoottime(ctx),
					IsDiagnostic:     true,
					IconString:       "mdi:restart",
					DeviceClassValue: types.DeviceClassTimestamp,
				},
			},
		},
		nil
}

func (w *timeWorker) getUptime(ctx context.Context) any {
	value, err := host.UptimeWithContext(ctx)
	if err != nil {
		w.logger.Debug("Failed to retrieve uptime.", "error", err.Error())

		return sensor.StateUnknown
	}

	epoch := time.Unix(0, 0)
	uptime := time.Unix(int64(value), 0)

	return uptime.Sub(epoch).Hours()
}

func (w *timeWorker) getBoottime(ctx context.Context) string {
	value, err := host.BootTimeWithContext(ctx)
	if err != nil {
		w.logger.Debug("Failed to retrieve boottime.", "error", err.Error())

		return sensor.StateUnknown
	}

	return time.Unix(int64(value), 0).Format(time.RFC3339)
}

func NewTimeWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &timeWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", timeWorkerID)),
			},
			WorkerID: timeWorkerID,
		},
		nil
}
