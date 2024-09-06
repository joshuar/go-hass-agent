// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package mem

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	updateInterval = time.Minute
	updateJitter   = 5 * time.Second

	memorySensorIcon = "mdi:memory"

	memoryUsageSensorUnits   = "B"
	memoryUsageSensorPcUnits = "%"

	workerID = "memory_usage_sensors"
)

// Lists of the memory statistics we want to track as sensors. See /proc/meminfo
// for all possible statistics.
var (
	memSensors  = []memStatID{memTotal, memFree, memBuffered, memCached, memAvailable, memCorrupted}
	swapSensors = []memStatID{swapTotal, swapFree, swapCached}
)

// newMemSensor generates a memorySensor for a memory stat.
func newMemSensor(id memStatID, stat *memStat) *linux.Sensor {
	var value uint64

	if stat == nil {
		value = 0
	} else {
		value = stat.value
	}

	return &linux.Sensor{
		DisplayName:      id.String(),
		UniqueID:         strcase.ToSnake(id.String()),
		DeviceClassValue: types.DeviceClassDataSize,
		StateClassValue:  types.StateClassTotal,
		IconString:       memorySensorIcon,
		DataSource:       linux.DataSrcProcfs,
		Value:            value,
		UnitsString:      memoryUsageSensorUnits,
	}
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(name string, value, total uint64) *linux.Sensor {
	var valuePc float64
	if total == 0 {
		valuePc = 0
	} else {
		valuePc = math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd
	}

	return &linux.Sensor{
		DisplayName:     name,
		UniqueID:        strcase.ToSnake(name),
		StateClassValue: types.StateClassMeasurement,
		IconString:      memorySensorIcon,
		DataSource:      linux.DataSrcProcfs,
		Value:           valuePc,
		UnitsString:     memoryUsageSensorPcUnits,
	}
}

// Calculate used memory = total - free/buffered/cached.
func newMemUsedPc(stats memoryStats) *linux.Sensor {
	var memOther uint64

	for name, stat := range stats {
		switch name {
		case memFree:
			memOther += stat.value
		case memBuffered:
			memOther += stat.value
		case memCached:
			memOther += stat.value
		}
	}

	memTotal, _ := stats.get(memTotal)

	memUsed := memTotal - memOther

	return newMemSensorPc("Memory Usage", memUsed, memTotal)
}

// Calculate used swap = total - free.
func newSwapUsedPc(stats memoryStats) *linux.Sensor {
	swapTotal, _ := stats.get(swapTotal)
	swapFree, _ := stats.get(swapFree)
	swapUsed := swapTotal - swapFree

	return newMemSensorPc("Swap Usage", swapUsed, swapTotal)
}

type usageWorker struct {
	logger *slog.Logger
}

func (w *usageWorker) Interval() time.Duration { return updateInterval }

func (w *usageWorker) Jitter() time.Duration { return updateJitter }

func (w *usageWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	stats, err := getMemStats()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve memory stats: %w", err)
	}

	sensors := make([]sensor.Details, 0, len(memSensors)+len(swapSensors)+2) //nolint:mnd

	for _, id := range memSensors {
		sensors = append(sensors, newMemSensor(id, stats[id]))
	}

	sensors = append(sensors, newMemUsedPc(stats))

	if stat, _ := stats.get(swapTotal); stat > 0 {
		for _, id := range swapSensors {
			sensors = append(sensors, newMemSensor(id, stats[id]))
		}

		sensors = append(sensors, newSwapUsedPc(stats))
	}

	return sensors, nil
}

func NewUsageWorker(ctx context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &usageWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", workerID)),
			},
			WorkerID: workerID,
		},
		nil
}
