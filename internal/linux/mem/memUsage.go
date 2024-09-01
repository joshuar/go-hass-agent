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

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
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
	return &linux.Sensor{
		DisplayName:      id.String(),
		DeviceClassValue: types.DeviceClassDataSize,
		StateClassValue:  types.StateClassTotal,
		IconString:       memorySensorIcon,
		DataSource:       linux.DataSrcProcfs,
		Value:            stat.value,
		UnitsString:      memoryUsageSensorUnits,
	}
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(name string, value, total uint64) *linux.Sensor {
	valuePc := math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd

	return &linux.Sensor{
		DisplayName:     name,
		StateClassValue: types.StateClassMeasurement,
		IconString:      memorySensorIcon,
		DataSource:      linux.DataSrcProcfs,
		Value:           valuePc,
		UnitsString:     memoryUsageSensorPcUnits,
	}
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

	sensors := make([]sensor.Details, 0, 9) //nolint:mnd

	for _, id := range memSensors {
		sensors = append(sensors, newMemSensor(id, stats[id]))
	}

	// Calculate used memory = total - free/buffered/cached.
	usedMem := stats[memTotal].value - stats[memFree].value - stats[memBuffered].value - stats[memCached].value
	sensors = append(sensors, newMemSensorPc("Memory Usage", usedMem, stats[memTotal].value))

	if stats[swapTotal].value > 0 {
		for _, id := range swapSensors {
			sensors = append(sensors, newMemSensor(id, stats[id]))
		}

		// Calculate used swap = total - free
		usedSwap := stats[swapTotal].value - stats[swapFree].value
		sensors = append(sensors, newMemSensorPc("Swap Usage", usedSwap, stats[swapTotal].value))
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
