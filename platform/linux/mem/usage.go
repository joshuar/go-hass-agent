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
	"slices"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	memUsageUpdateInterval = time.Minute
	memUsageUpdateJitter   = 5 * time.Second

	memorySensorIcon = "mdi:memory"

	memoryUsageSensorUnits   = "B"
	memoryUsageSensorPcUnits = "%"

	defaultGPUVendor = GPUVendorNone
	defaultGPUCard   = "card1"
)

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
)

// Lists of the memory statistics we want to track as sensors. See /proc/meminfo
// for all possible statistics.
// For (amd) gpus, more statistics are available at /sys/class/drm/cardX/device/
var (
	gpuMemSensors = []gpuMemStatID{memVRamTotal, memGTTTotal, memVRamUsed, memGTTUsed}
	memSensors    = []memStatID{memTotal, memFree, memBuffered, memCached, memAvailable, memCorrupted}
	swapSensors   = []memStatID{swapTotal, swapFree, swapCached}
)

// newMemSensor generates a memorySensor for a memory stat.
func newMemSensor(ctx context.Context, id memStatID, stat *memStat) models.Entity {
	var value uint64

	if stat == nil {
		value = 0
	} else {
		value = stat.value
	}

	return models.NewSensor(ctx,
		models.WithName(id.String()),
		models.WithID(strcase.ToSnake(id.String())),
		models.WithUnits(memoryUsageSensorUnits),
		models.WithDeviceClass(models.SensorClassDataSize),
		models.WithStateClass(models.StateTotal),
		models.WithIcon(memorySensorIcon),
		models.WithState(value),
		models.WithDataSourceAttribute(linux.DataSrcProcFS),
		models.WithAttribute("native_unit_of_measurement", memoryUsageSensorUnits),
	)
}

// newGpuMemSensor generates a memorySensor for a gpu memory stat.
func newGpuMemSensor(ctx context.Context, id gpuMemStatID, stat uint64) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(id.String()),
		models.WithID(strcase.ToSnake(id.String())),
		models.WithUnits(memoryUsageSensorUnits),
		models.WithDeviceClass(models.SensorClassDataSize),
		models.WithStateClass(models.StateTotal),
		models.WithIcon(memorySensorIcon),
		models.WithState(stat),
		models.WithDataSourceAttribute(linux.DataSrcSysFS),
		models.WithAttribute("native_unit_of_measurement", memoryUsageSensorUnits),
	)
}

// newGpuMemSensorPc generates a gpu memorySensor with a percentage value for a memory
// stat.
func newGpuMemSensorPc(ctx context.Context, name string, value, total uint64) models.Entity {
	var valuePc float64
	if total == 0 {
		valuePc = 0
	} else {
		valuePc = math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd // %
	}

	return models.NewSensor(ctx,
		models.WithName(name),
		models.WithID(strcase.ToSnake(name)),
		models.WithUnits(memoryUsageSensorPcUnits),
		models.WithStateClass(models.StateTotal),
		models.WithIcon(memorySensorIcon),
		models.WithState(valuePc),
		models.WithDataSourceAttribute(linux.DataSrcSysFS),
		models.WithAttribute("native_unit_of_measurement", memoryUsageSensorPcUnits),
	)
}

// newMemSensorPc generates a memorySensor with a percentage value for a memory
// stat.
func newMemSensorPc(ctx context.Context, name string, value, total uint64) models.Entity {
	var valuePc float64
	if total == 0 {
		valuePc = 0
	} else {
		valuePc = math.Round(float64(value)/float64(total)*100/0.05) * 0.05 //nolint:mnd // %
	}

	return models.NewSensor(ctx,
		models.WithName(name),
		models.WithID(strcase.ToSnake(name)),
		models.WithUnits(memoryUsageSensorPcUnits),
		models.WithStateClass(models.StateTotal),
		models.WithIcon(memorySensorIcon),
		models.WithState(valuePc),
		models.WithDataSourceAttribute(linux.DataSrcProcFS),
		models.WithAttribute("native_unit_of_measurement", memoryUsageSensorPcUnits),
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
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *WorkerPreferences
}

func NewUsageWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata("mem_usage", "Memory usage"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &WorkerPreferences{
		UpdateInterval: memUsageUpdateInterval.String(),
		GPUVendor:      defaultGPUVendor.String(),
		GPUCard:        defaultGPUCard,
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(prefPrefix+"usage", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = memUsageUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, memUsageUpdateJitter)

	return worker, nil
}

func (w *usageWorker) Execute(ctx context.Context) error {
	var (
		stats    memoryStats
		gpuStats gpuMemoryStats
		err      error
	)

	ctx = slogctx.With(ctx, "worker", w.ID())

	stats, err = getMemStats(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve memory stats: %w", err)
	}

	// Memory sensors.
	for stat := range slices.Values(memSensors) {
		w.OutCh <- newMemSensor(ctx, stat, stats[stat])
	}
	w.OutCh <- newMemUsedPc(ctx, stats)

	if gpuVendorTypeNames[w.prefs.GPUVendor] == GPUVendorAMD {
		// (AMD) GPU memory sensors
		gpuStats, err = getAmdGpuMemStats(ctx, w.prefs.GPUCard)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Get AMD gpu stats failed.",
				slog.Any("error", err))
		} else {
			for gpuStat := range slices.Values(gpuMemSensors) {
				w.OutCh <- newGpuMemSensor(ctx, gpuStat, gpuStats[gpuStat])
			}
			w.OutCh <- newGpuMemSensorPc(ctx, "GPU VRam Usage", gpuStats[memVRamUsed], gpuStats[memVRamTotal])
			w.OutCh <- newGpuMemSensorPc(ctx, "GPU GTT Usage", gpuStats[memGTTUsed], gpuStats[memGTTTotal])
		}
	}

	// Swap memory sensors.
	if stat, _ := stats.get(swapTotal); stat > 0 {
		for id := range slices.Values(swapSensors) {
			w.OutCh <- newMemSensor(ctx, id, stats[id])
		}
		w.OutCh <- newSwapUsedPc(ctx, stats)
	}

	return nil
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
