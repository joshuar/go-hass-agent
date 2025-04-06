// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	ioWorkerUpdateInterval = 5 * time.Second
	ioWorkerUpdateJitter   = time.Second

	ioWorkerID   = "disk_rates_sensors"
	ioWorkerDesc = "IO usage stats"

	totalsID = "total"
)

var ErrInitRatesWorker = errors.New("could not init rates worker")

var (
	_ quartz.Job                  = (*ioWorker)(nil)
	_ workers.PollingEntityWorker = (*ioWorker)(nil)
)

// ioWorker creates sensors for disk IO counts and rates per device. It
// maintains an internal map of devices being tracked.
type ioWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
	boottime    time.Time
	rateSensors map[string]map[ioSensor]*ioRate
	mu          sync.Mutex
	prefs       *WorkerPrefs
}

// addDevice adds a new device to the tracker map. If sthe device is already
// being tracked, it will not be added again. The bool return indicates whether
// a device was added (true) or not (false).
func (w *ioWorker) addRateSensors(dev *device) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.rateSensors[dev.id]; !found {
		w.rateSensors[dev.id] = map[ioSensor]*ioRate{
			diskReadRate:  {rateType: diskReadRate},
			diskWriteRate: {rateType: diskWriteRate},
		}
	}
}

func (w *ioWorker) generateDeviceRateSensors(ctx context.Context, device *device, stats map[stat]uint64, delta time.Duration) ([]models.Entity, error) {
	var (
		sensors  []models.Entity
		warnings error
	)

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.rateSensors[device.id]; found && stats != nil {
		for rateType := range w.rateSensors[device.id] {
			var currValue uint64

			switch rateType {
			case diskReadRate:
				currValue = stats[TotalSectorsRead]
			case diskWriteRate:
				currValue = stats[TotalSectorsWritten]
			}

			rate := w.rateSensors[device.id][rateType].Calculate(currValue, delta)

			entity, err := newDiskRateSensor(ctx, device, rateType, rate)
			if err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("could not generate rate sensor: %w", err))
			} else {
				sensors = append(sensors, *entity)
			}
		}
	}

	return sensors, warnings
}

func (w *ioWorker) generateDeviceStatSensors(ctx context.Context, device *device, stats map[stat]uint64) ([]models.Entity, error) {
	var (
		sensors  []models.Entity
		entity   *models.Entity
		err      error
		warnings error
	)

	diskReadsAttributes := models.Attributes{
		"total_sectors_read":         stats[TotalSectorsRead],
		"total_milliseconds_reading": stats[TotalTimeReading],
	}

	diskWriteAttributes := models.Attributes{
		"total_sectors_written":      stats[TotalSectorsWritten],
		"total_milliseconds_writing": stats[TotalTimeWriting],
	}

	// Generate diskReads sensor for device.
	entity, err = newDiskStatSensor(ctx, device, diskReads, stats[TotalReads], diskReadsAttributes)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("create disk read stats failed: %w", err))
	} else {
		sensors = append(sensors, *entity)
	}
	// Generate diskWrites sensor for device.
	entity, err = newDiskStatSensor(ctx, device, diskWrites, stats[TotalWrites], diskWriteAttributes)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("create disk write stats failed: %w", err))
	} else {
		sensors = append(sensors, *entity)
	}
	// Generate IOsInProgress sensor for device.
	entity, err = newDiskStatSensor(ctx, device, diskIOInProgress, stats[ActiveIOs], nil)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("create IOs in progress failed: %w", err))
	} else {
		sensors = append(sensors, *entity)
	}

	return sensors, warnings
}

func (w *ioWorker) Execute(ctx context.Context) error {
	delta := w.GetDelta()
	// Get valid devices.
	deviceNames, err := getDeviceNames()
	if err != nil {
		return fmt.Errorf("could not fetch disk devices: %w", err)
	}

	statsTotals := make(map[stat]uint64)

	// Get the current device info and stats for all valid devices.
	for dev := range slices.Values(deviceNames) {
		dev, stats, err := getDevice(dev)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", ioWorkerID)).
				Debug("Unable to read device stats.", slog.Any("error", err))

			continue
		}

		// Add rate sensors for device (if not already added).
		w.addRateSensors(dev)

		rateSensors, warnings := w.generateDeviceRateSensors(ctx, dev, stats, delta)
		if warnings != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", ioWorkerID)).
				Debug("Some problems occurred generating disk rate sensors.", slog.Any("warnings", warnings))
		}
		for s := range slices.Values(rateSensors) {
			w.OutCh <- s
		}

		statSensors, warnings := w.generateDeviceStatSensors(ctx, dev, stats)
		if warnings != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", ioWorkerID)).
				Debug("Some problems occurred generating disk rate sensors.", slog.Any("warnings", warnings))
		}
		for s := range slices.Values(statSensors) {
			w.OutCh <- s
		}

		// Don't include "aggregate" devices in totals.
		if strings.HasPrefix(dev.id, "dm") || strings.HasPrefix(dev.id, "md") {
			continue
		}
		// Add device stats to the totals.
		for stat, value := range stats {
			statsTotals[stat] += value
		}
	}

	// Update total stats.
	rateSensors, warnings := w.generateDeviceRateSensors(ctx, &device{id: totalsID}, statsTotals, delta)
	if warnings != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", ioWorkerID)).
			Debug("Some problems occurred generating disk rate sensors.", slog.Any("warnings", warnings))
	}
	for s := range slices.Values(rateSensors) {
		w.OutCh <- s
	}

	statSensors, warnings := w.generateDeviceStatSensors(ctx, &device{id: totalsID}, statsTotals)
	if warnings != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", ioWorkerID)).
			Debug("Some problems occurred generating disk rate sensors.", slog.Any("warnings", warnings))
	}
	for s := range slices.Values(statSensors) {
		w.OutCh <- s
	}

	return nil
}

func (w *ioWorker) PreferencesID() string {
	return ioWorkerPreferencesID
}

func (w *ioWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		UpdateInterval: ioWorkerUpdateInterval.String(),
	}
}

func (w *ioWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ioWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

func NewIOWorker(ctx context.Context) (workers.EntityWorker, error) {
	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, errors.Join(ErrInitRatesWorker,
			fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx))
	}

	// Add sensors for a pseudo "total" device which tracks total values from
	// all devices.
	devices := make(map[string]map[ioSensor]*ioRate)
	devices["total"] = map[ioSensor]*ioRate{
		diskReadRate:  {rateType: diskReadRate},
		diskWriteRate: {rateType: diskWriteRate},
	}

	ioWorker := &ioWorker{
		WorkerMetadata:          models.SetWorkerMetadata(ioWorkerID, ioWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		rateSensors:             devices,
		boottime:                boottime,
	}

	prefs, err := preferences.LoadWorker(ioWorker)
	if err != nil {
		return nil, errors.Join(ErrInitRatesWorker, err)
	}
	ioWorker.prefs = prefs

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", ioWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", ioWorkerUpdateInterval.String()))

		pollInterval = ioWorkerUpdateInterval
	}
	ioWorker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, ioWorkerUpdateJitter)

	return ioWorker, nil
}
