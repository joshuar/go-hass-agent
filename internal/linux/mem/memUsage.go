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
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	scaleFactor = 100

	updateInterval = time.Minute
	updateJitter   = 5 * time.Second
)

var ErrUnknownSensor = errors.New("unknown sensor")

type memorySensor struct {
	linux.Sensor
}

func (s *memorySensor) Attributes() any {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"data_source"`
	}{
		NativeUnit: s.UnitsString,
		DataSource: s.SensorSrc,
	}
}

//nolint:exhaustive,exhaustruct
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

type usageWorker struct{}

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
			log.Warn().Err(err).Msg("Could not retrieve memory usage stat.")

			continue
		}

		sensors = append(sensors, memSensor)
	}

	return sensors, nil
}

func NewUsageWorker(_ context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Memory Usage Sensor",
			WorkerDesc: "System RAM (and swap if enabled) usage as a percentage.",
			Value:      &usageWorker{},
		},
		nil
}
