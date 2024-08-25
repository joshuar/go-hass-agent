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
	"math"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	updateInterval = time.Minute
	updateJitter   = 5 * time.Second

	workerID = "memory_usage_sensors"
)

var ErrUnknownSensor = errors.New("unknown sensor")

var (
	memSensors  = []memStatID{memTotal, memFree, memBuffered, memCached, memAvailable, memCorrupted}
	swapSensors = []memStatID{swapTotal, swapFree, swapCached}
)

type memorySensor struct {
	name string
	linux.Sensor
}

func (s *memorySensor) Name() string {
	return s.name
}

func (s *memorySensor) ID() string {
	return strcase.ToSnake(s.name)
}

// newMemSensor generates a memorySensor for a memory stat.
func newMemSensor(id memStatID, stat *memStat) *memorySensor {
	return &memorySensor{
		Sensor: linux.Sensor{
			SensorTypeValue:  linux.SensorMemoryStat,
			DeviceClassValue: types.DeviceClassDataSize,
			StateClassValue:  types.StateClassTotal,
			IconString:       "mdi:memory",
			SensorSrc:        linux.DataSrcProcfs,
			Value:            stat.value,
			UnitsString:      "B	",
		},
		name: id.String(),
	}
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(name string, value, total uint64) *memorySensor {
	valuePc := math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd

	return &memorySensor{
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorMemoryStat,
			StateClassValue: types.StateClassMeasurement,
			IconString:      "mdi:memory",
			SensorSrc:       linux.DataSrcProcfs,
			Value:           valuePc,
			UnitsString:     "%",
		},
		name: name,
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
