// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

type memorySensor struct {
	linux.Sensor
}

func (s *memorySensor) Attributes() any {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
	}{
		NativeUnit: s.UnitsString,
		DataSource: s.SensorSrc,
	}
}

func newMemoryUsageSensor(sensorType linux.SensorTypeValue, stats *mem.VirtualMemoryStat) (*memorySensor, error) {
	s := &memorySensor{}

	s.SensorTypeValue = sensorType
	s.IconString = "mdi:memory"
	s.SensorSrc = linux.DataSrcSysfs

	switch s.SensorTypeValue {
	case linux.SensorMemTotal:
		s.Value = stats.Total
		s.UnitsString = "B"
		s.DeviceClassValue = types.DeviceClassDataSize
		s.StateClassValue = types.StateClassTotal
	case linux.SensorMemAvail:
		s.Value = stats.Available
		s.UnitsString = "B"
		s.DeviceClassValue = types.DeviceClassDataSize
		s.StateClassValue = types.StateClassTotal
	case linux.SensorMemUsed:
		s.Value = stats.Used
		s.UnitsString = "B"
		s.DeviceClassValue = types.DeviceClassDataSize
		s.StateClassValue = types.StateClassTotal
	case linux.SensorMemPc:
		s.Value = float64(stats.Used) / float64(stats.Total) * 100
		s.UnitsString = "%"
		s.StateClassValue = types.StateClassMeasurement
	case linux.SensorSwapTotal:
		s.Value = stats.SwapTotal
		s.UnitsString = "B"
		s.DeviceClassValue = types.DeviceClassDataSize
		s.StateClassValue = types.StateClassTotal
	case linux.SensorSwapFree:
		s.Value = stats.SwapFree
		s.UnitsString = "B"
		s.DeviceClassValue = types.DeviceClassDataSize
		s.StateClassValue = types.StateClassTotal
	case linux.SensorSwapPc:
		s.Value = float64(stats.SwapCached) / float64(stats.SwapTotal) * 100
		s.UnitsString = "%"
		s.StateClassValue = types.StateClassMeasurement
	default:
		return nil, errors.New("unknown memory stat")
	}
	return s, nil
}

type usageWorker struct{}

func (w *usageWorker) Interval() time.Duration { return time.Minute }

func (w *usageWorker) Jitter() time.Duration { return 5 * time.Second }

func (w *usageWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	var sensors []sensor.Details
	var memDetails *mem.VirtualMemoryStat
	var err error
	if memDetails, err = mem.VirtualMemoryWithContext(ctx); err != nil {
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

	for _, stat := range stats {
		s, err := newMemoryUsageSensor(stat, memDetails)
		if err != nil {
			log.Warn().Err(err).Msg("Could not retrieve memory usage stat.")
			continue
		}
		sensors = append(sensors, s)
	}
	return sensors, nil
}

func NewUsageWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Memory Usage Sensor",
			WorkerDesc: "System RAM (and swap if enabled) usage as a percentage.",
			Value:      &usageWorker{},
		},
		nil
}
