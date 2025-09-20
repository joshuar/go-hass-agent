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
	"slices"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	memUsageUpdateInterval = time.Minute
	memUsageUpdateJitter   = 5 * time.Second

	memorySensorIcon = "mdi:memory"

	memoryUsageSensorUnits   = "B"
	memoryUsageSensorPcUnits = "%"

	memUsageWorkerID      = "memory_usage_sensors"
	memUsageWorkerDesc    = "Memory usage stats"
	memUsagePreferencesID = prefPrefix + "usage"
)

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
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
func newMemSensor(ctx context.Context, id memStatID, stat *memStat) models.Entity {
	var value uint64

	if stat == nil {
		value = 0
	} else {
		value = stat.value
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(id.String()),
		sensor.WithID(strcase.ToSnake(id.String())),
		sensor.WithUnits(memoryUsageSensorUnits),
		sensor.WithDeviceClass(class.SensorClassDataSize),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(memorySensorIcon),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcProcFS),
		sensor.WithAttribute("native_unit_of_measurement", memoryUsageSensorUnits),
	)
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(ctx context.Context, name string, value, total uint64) models.Entity {
	var valuePc float64
	if total == 0 {
		valuePc = 0
	} else {
		valuePc = math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithUnits(memoryUsageSensorPcUnits),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(memorySensorIcon),
		sensor.WithState(valuePc),
		sensor.WithDataSourceAttribute(linux.DataSrcProcFS),
		sensor.WithAttribute("native_unit_of_measurement", memoryUsageSensorPcUnits),
	)
}

// Calculate used memory = total - free/buffered/cached.
func newMemUsedPc(ctx context.Context, stats memoryStats) models.Entity {
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
func newSwapUsedPc(ctx context.Context, stats memoryStats) models.Entity {
	swapTotal, _ := stats.get(swapTotal)
	swapFree, _ := stats.get(swapFree)
	swapUsed := swapTotal - swapFree

	return newMemSensorPc(ctx, "Swap Usage", swapUsed, swapTotal)
}

type usageWorker struct {
	prefs *WorkerPreferences
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
}

func (w *usageWorker) Execute(ctx context.Context) error {
	var (
		stats memoryStats
		err   error
	)

	stats, err = getMemStats()
	if err != nil {
		return fmt.Errorf("unable to retrieve memory stats: %w", err)
	}

	// Memory sensors.
	for stat := range slices.Values(memSensors) {
		w.OutCh <- newMemSensor(ctx, stat, stats[stat])
	}
	w.OutCh <- newMemUsedPc(ctx, stats)

	// Swap memory sensors.
	if stat, _ := stats.get(swapTotal); stat > 0 {
		for _, id := range swapSensors {
			w.OutCh <- newMemSensor(ctx, id, stats[id])
		}
		w.OutCh <- newSwapUsedPc(ctx, stats)
	}

	return nil
}

func (w *usageWorker) PreferencesID() string {
	return memUsagePreferencesID
}

func (w *usageWorker) DefaultPreferences() WorkerPreferences {
	return WorkerPreferences{
		UpdateInterval: memUsageUpdateInterval.String(),
	}
}

func (w *usageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *usageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func NewUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata(memUsageWorkerID, memUsageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &WorkerPreferences{
		UpdateInterval: memUsageUpdateInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(memUsagePreferencesID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", memUsageWorkerID),
			slog.String("given_interval", worker.prefs.UpdateInterval),
			slog.String("default_interval", memUsageUpdateInterval.String()))

		pollInterval = memUsageUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, memUsageUpdateJitter)

	return worker, nil
}
