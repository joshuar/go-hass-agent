// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package disk

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	ratesUpdateInterval = 5 * time.Second
	ratesUpdateJitter   = time.Second

	ratesWorkerID = "disk_rates_sensors"

	totalsID = "total"
)

// ioWorker creates sensors for disk IO counts and rates per device. It
// maintains an internal map of devices being tracked.
type ioWorker struct {
	boottime time.Time
	devices  map[string][]*diskIOSensor
	linux.PollingSensorWorker
	delta time.Duration
	mu    sync.Mutex
}

// addDevice adds a new device to the tracker map. If sthe device is already
// being tracked, it will not be added again. The bool return indicates whether
// a device was added (true) or not (false).
func (w *ioWorker) addDevice(dev *device) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.devices[dev.id]; !found {
		w.devices[dev.id] = newDeviceSensors(w.boottime, dev)
	}
}

// updateDevice will update a tracked device's stats. For rates, it will
// recalculate based on the given time delta.
func (w *ioWorker) updateDevice(id string, stats map[stat]uint64, delta time.Duration) []sensor.Entity {
	w.mu.Lock()
	defer w.mu.Unlock()

	sensors := make([]sensor.Entity, len(w.devices[id]))

	if _, found := w.devices[id]; found && stats != nil {
		for idx := range w.devices[id] {
			w.devices[id][idx].update(stats, delta)
			sensors[idx] = *w.devices[id][idx].Entity
		}
	}

	return sensors
}

func (w *ioWorker) UpdateDelta(delta time.Duration) {
	w.delta = delta
}

func (w *ioWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	// Get valid devices.
	deviceNames, err := getDeviceNames()
	if err != nil {
		return nil, fmt.Errorf("could not fetch disk devices: %w", err)
	}

	sensors := make([]sensor.Entity, 0, 4*len(deviceNames)+4) //nolint:mnd
	statsTotals := make(map[stat]uint64)

	// Get the current device info and stats for all valid devices.
	for _, name := range deviceNames {
		dev, stats, err := getDevice(name)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", ratesWorkerID)).
				Debug("Unable to read device stats.", slog.Any("error", err))

			continue
		}

		// Add device (if it isn't already tracked).
		w.addDevice(dev)

		// Update device stats and return updated sensors.
		sensors = append(sensors, w.updateDevice(dev.id, stats, w.delta)...)

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
	sensors = append(sensors, w.updateDevice(totalsID, statsTotals, w.delta)...)

	return sensors, nil
}

func NewIOWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(ratesWorkerID, ratesUpdateInterval, ratesUpdateJitter)

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	// Add sensors for a pseudo "total" device which tracks total values from
	// all devices.
	devices := make(map[string][]*diskIOSensor)
	devices["total"] = newDeviceSensors(boottime, &device{id: totalsID})

	worker.PollingType = &ioWorker{
		devices:  devices,
		boottime: boottime,
	}

	return worker, nil
}

func newDeviceSensors(boottime time.Time, dev *device) []*diskIOSensor {
	return []*diskIOSensor{
		newDiskIOSensor(boottime, dev, diskReads),
		newDiskIOSensor(boottime, dev, diskWrites),
		newDiskIORateSensor(dev, diskReadRate),
		newDiskIORateSensor(dev, diskWriteRate),
	}
}
