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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	memUsageUpdateInterval = time.Minute
	memUsageUpdateJitter   = 5 * time.Second

	memorySensorIcon = "mdi:memory"

	memoryUsageSensorUnits   = "B"
	memoryUsageSensorPcUnits = "%"

	memUsageWorkerID      = "memory_usage_sensors"
	memUsagePreferencesID = prefPrefix + "usage"
)

var ErrNewMemStatSensor = errors.New("could not create mem stat sensor")

// Lists of the memory statistics we want to track as sensors. See /proc/meminfo
// for all possible statistics.
var (
	memSensors  = []memStatID{memTotal, memFree, memBuffered, memCached, memAvailable, memCorrupted}
	swapSensors = []memStatID{swapTotal, swapFree, swapCached}
)

var ErrInitUsageWorker = errors.New("could not init memory usage worker")

// newMemSensor generates a memorySensor for a memory stat.
func newMemSensor(ctx context.Context, id memStatID, stat *memStat) (*models.Entity, error) {
	var value uint64

	if stat == nil {
		value = 0
	} else {
		value = stat.value
	}

	statSensor, err := sensor.NewSensor(ctx,
		sensor.WithName(id.String()),
		sensor.WithID(strcase.ToSnake(id.String())),
		sensor.WithUnits(memoryUsageSensorUnits),
		sensor.WithDeviceClass(class.SensorClassDataSize),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(memorySensorIcon),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		sensor.WithAttribute("native_unit_of_measurement", memoryUsageSensorUnits),
	)
	if err != nil {
		return nil, errors.Join(ErrNewMemStatSensor, err)
	}

	return &statSensor, nil
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(ctx context.Context, name string, value, total uint64) (*models.Entity, error) {
	var valuePc float64
	if total == 0 {
		valuePc = 0
	} else {
		valuePc = math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd
	}

	statSensor, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithUnits(memoryUsageSensorPcUnits),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(memorySensorIcon),
		sensor.WithState(valuePc),
		sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		sensor.WithAttribute("native_unit_of_measurement", memoryUsageSensorPcUnits),
	)
	if err != nil {
		return nil, errors.Join(ErrNewMemStatSensor, err)
	}

	return &statSensor, nil
}

// Calculate used memory = total - free/buffered/cached.
func newMemUsedPc(ctx context.Context, stats memoryStats) (*models.Entity, error) {
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

	return newMemSensorPc(ctx, "Memory Usage", memUsed, memTotal)
}

// Calculate used swap = total - free.
func newSwapUsedPc(ctx context.Context, stats memoryStats) (*models.Entity, error) {
	swapTotal, _ := stats.get(swapTotal)
	swapFree, _ := stats.get(swapFree)
	swapUsed := swapTotal - swapFree

	return newMemSensorPc(ctx, "Swap Usage", swapUsed, swapTotal)
}

type usageWorker struct{}

func (w *usageWorker) UpdateDelta(_ time.Duration) {}

func (w *usageWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	var (
		stats  memoryStats
		entity *models.Entity
		err    error
	)

	stats, err = getMemStats()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve memory stats: %w", err)
	}

	sensors := make([]models.Entity, 0, len(memSensors)+len(swapSensors)+2) //nolint:mnd

	for _, id := range memSensors {
		entity, err = newMemSensor(ctx, id, stats[id])
		if err != nil {
			logging.FromContext(ctx).Warn("Could not generate memory usage sensor.", slog.Any("error", err))
			continue
		}

		sensors = append(sensors, *entity)
	}

	entity, err = newMemUsedPc(ctx, stats)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not generate memory usage sensor.", slog.Any("error", err))
	} else {
		sensors = append(sensors, *entity)
	}

	if stat, _ := stats.get(swapTotal); stat > 0 {
		for _, id := range swapSensors {
			entity, err := newMemSensor(ctx, id, stats[id])
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate swap usage sensor.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, *entity)
		}

		entity, err := newSwapUsedPc(ctx, stats)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not generate memory usage sensor.", slog.Any("error", err))
		} else {
			sensors = append(sensors, *entity)
		}
	}

	return sensors, nil
}

func (w *usageWorker) PreferencesID() string {
	return memUsagePreferencesID
}

func (w *usageWorker) DefaultPreferences() WorkerPreferences {
	return WorkerPreferences{
		UpdateInterval: memUsageUpdateInterval.String(),
	}
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	usageWorker := &usageWorker{}

	prefs, err := preferences.LoadWorker(usageWorker)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", memUsageWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", memUsageUpdateInterval.String()))

		pollInterval = memUsageUpdateInterval
	}

	worker := linux.NewPollingSensorWorker(memUsageWorkerID, pollInterval, memUsageUpdateJitter)
	worker.PollingSensorType = usageWorker

	return worker, nil
}
