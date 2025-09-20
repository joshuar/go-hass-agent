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
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
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

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: ioWorkerUpdateInterval.String(),
	}
	var err error
	ioWorker.prefs, err = workers.LoadWorkerPreferences(ioWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitRatesWorker, err)
	}

	pollInterval, err := time.ParseDuration(ioWorker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", ioWorkerID),
			slog.String("given_interval", ioWorker.prefs.UpdateInterval),
			slog.String("default_interval", ioWorkerUpdateInterval.String()))

		pollInterval = ioWorkerUpdateInterval
	}
	ioWorker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, ioWorkerUpdateJitter)

	return ioWorker, nil
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
			slogctx.FromCtx(ctx).
				With(slog.String("worker", ioWorkerID)).
				Debug("Unable to read device stats.", slog.Any("error", err))

			continue
		}

		// Add rate sensors for device (if not already added).
		w.addRateSensors(dev)
		for s := range slices.Values(w.generateDeviceRateSensors(ctx, dev, stats, delta)) {
			w.OutCh <- s
		}

		for s := range slices.Values(w.generateDeviceStatSensors(ctx, dev, stats)) {
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
	for s := range slices.Values(w.generateDeviceRateSensors(ctx, &device{id: totalsID}, statsTotals, delta)) {
		w.OutCh <- s
	}
	for s := range slices.Values(w.generateDeviceStatSensors(ctx, &device{id: totalsID}, statsTotals)) {
		w.OutCh <- s
	}

	return nil
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

func (w *ioWorker) generateDeviceRateSensors(ctx context.Context, device *device, stats map[stat]uint64, delta time.Duration) []models.Entity {
	var sensors []models.Entity

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
			sensors = append(sensors, newDiskRateSensor(ctx, device, rateType, rate))
		}
	}

	return sensors
}

func (w *ioWorker) generateDeviceStatSensors(ctx context.Context, device *device, stats map[stat]uint64) []models.Entity {
	var sensors []models.Entity

	diskReadsAttributes := models.Attributes{
		"total_sectors_read":         stats[TotalSectorsRead],
		"total_milliseconds_reading": stats[TotalTimeReading],
	}

	diskWriteAttributes := models.Attributes{
		"total_sectors_written":      stats[TotalSectorsWritten],
		"total_milliseconds_writing": stats[TotalTimeWriting],
	}

	// Generate diskReads sensor for device.
	sensors = append(sensors, newDiskStatSensor(ctx, device, diskReads, stats[TotalReads], diskReadsAttributes))
	// Generate diskWrites sensor for device.
	sensors = append(sensors, newDiskStatSensor(ctx, device, diskWrites, stats[TotalWrites], diskWriteAttributes))
	// Generate IOsInProgress sensor for device.
	sensors = append(sensors, newDiskStatSensor(ctx, device, diskIOInProgress, stats[ActiveIOs], nil))

	return sensors
}
