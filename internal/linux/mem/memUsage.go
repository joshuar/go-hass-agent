// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package mem

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
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
func newMemSensor(id memStatID, stat *memStat) sensor.Entity {
	var value uint64

	if stat == nil {
		value = 0
	} else {
		value = stat.value
	}

	return sensor.Entity{
		Name:        id.String(),
		DeviceClass: types.SensorDeviceClassDataSize,
		StateClass:  types.StateClassTotal,
		Units:       memoryUsageSensorUnits,
		EntityState: &sensor.EntityState{
			ID:    strcase.ToSnake(id.String()),
			Icon:  memorySensorIcon,
			State: value,
			Attributes: map[string]any{
				"data_source":                linux.DataSrcProcfs,
				"native_unit_of_measurement": memoryUsageSensorUnits,
			},
		},
	}
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(name string, value, total uint64) sensor.Entity {
	var valuePc float64
	if total == 0 {
		valuePc = 0
	} else {
		valuePc = math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd
	}

	return sensor.Entity{
		Name:       name,
		StateClass: types.StateClassTotal,
		Units:      memoryUsageSensorPcUnits,
		EntityState: &sensor.EntityState{
			ID:    strcase.ToSnake(name),
			Icon:  memorySensorIcon,
			State: valuePc,
			Attributes: map[string]any{
				"data_source": linux.DataSrcProcfs,
			},
		},
	}
}

// Calculate used memory = total - free/buffered/cached.
func newMemUsedPc(stats memoryStats) sensor.Entity {
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
func newSwapUsedPc(stats memoryStats) sensor.Entity {
	swapTotal, _ := stats.get(swapTotal)
	swapFree, _ := stats.get(swapFree)
	swapUsed := swapTotal - swapFree

	return newMemSensorPc("Swap Usage", swapUsed, swapTotal)
}

type usageWorker struct{}

func (w *usageWorker) UpdateDelta(_ time.Duration) {}

func (w *usageWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	stats, err := getMemStats()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve memory stats: %w", err)
	}

	sensors := make([]sensor.Entity, 0, len(memSensors)+len(swapSensors)+2) //nolint:mnd

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

func NewUsageWorker(_ context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(workerID, updateInterval, updateJitter)
	worker.PollingType = &usageWorker{}

	return worker, nil
}
