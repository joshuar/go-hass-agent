// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package mem

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	scaleFactor = 100

	updateInterval = time.Minute
	updateJitter   = 5 * time.Second

	workerID = "memory_usage_sensors"
)

var ErrUnknownSensor = errors.New("unknown sensor")

type memorySensor struct {
	linux.Sensor
}

//nolint:exhaustive
func newMemoryUsageSensor(sensorType linux.SensorTypeValue, stats *mem.VirtualMemoryStat) (*memorySensor, error) {
	newSensor := &memorySensor{}

	newSensor.SensorTypeValue = sensorType
	newSensor.IconString = "mdi:memory"
	newSensor.SensorSrc = linux.DataSrcSysfs

	switch newSensor.SensorTypeValue {
	case linux.SensorMemTotal:
		newSensor.Value = stats.Total
		newSensor.UnitsString = "B"
		newSensor.DeviceClassValue = types.DeviceClassDataSize
		newSensor.StateClassValue = types.StateClassTotal
	case linux.SensorMemAvail:
		newSensor.Value = stats.Available
		newSensor.UnitsString = "B"
		newSensor.DeviceClassValue = types.DeviceClassDataSize
		newSensor.StateClassValue = types.StateClassTotal
	case linux.SensorMemUsed:
		newSensor.Value = stats.Used
		newSensor.UnitsString = "B"
		newSensor.DeviceClassValue = types.DeviceClassDataSize
		newSensor.StateClassValue = types.StateClassTotal
	case linux.SensorMemPc:
		newSensor.Value = float64(stats.Used) / float64(stats.Total) * scaleFactor
		newSensor.UnitsString = "%"
		newSensor.StateClassValue = types.StateClassMeasurement
	case linux.SensorSwapTotal:
		newSensor.Value = stats.SwapTotal
		newSensor.UnitsString = "B"
		newSensor.DeviceClassValue = types.DeviceClassDataSize
		newSensor.StateClassValue = types.StateClassTotal
	case linux.SensorSwapFree:
		newSensor.Value = stats.SwapFree
		newSensor.UnitsString = "B"
		newSensor.DeviceClassValue = types.DeviceClassDataSize
		newSensor.StateClassValue = types.StateClassTotal
	case linux.SensorSwapPc:
		newSensor.Value = float64(stats.SwapCached) / float64(stats.SwapTotal) * scaleFactor
		newSensor.UnitsString = "%"
		newSensor.StateClassValue = types.StateClassMeasurement
	default:
		return nil, ErrUnknownSensor
	}

	return newSensor, nil
}

type usageWorker struct {
	logger *slog.Logger
}

func (w *usageWorker) Interval() time.Duration { return updateInterval }

func (w *usageWorker) Jitter() time.Duration { return updateJitter }

func (w *usageWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	memDetails, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("problem fetching memory stats: %w", err)
	}

	// Memory stats to track as sensors.
	stats := []linux.SensorTypeValue{
		linux.SensorMemTotal,
		linux.SensorMemAvail,
		linux.SensorMemUsed,
		linux.SensorMemPc,
	}
	// If this system has swap enabled, track swap stats as sensors as well.
	if memDetails.SwapTotal > 0 {
		stats = append(stats,
			linux.SensorSwapTotal,
			linux.SensorSwapFree,
			linux.SensorSwapPc,
		)
	}

	sensors := make([]sensor.Details, 0, len(stats))

	for _, stat := range stats {
		memSensor, err := newMemoryUsageSensor(stat, memDetails)
		if err != nil {
			w.logger.Warn("Could not retrieve memory usage stats.", "error", err.Error())

			continue
		}

		sensors = append(sensors, memSensor)
	}

	return sensors, nil
}

func NewUsageWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &usageWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", workerID)),
			},
			WorkerID: workerID,
		},
		nil
}
