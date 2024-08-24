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
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	ratesUpdateInterval = 5 * time.Second
	ratesUpdateJitter   = time.Second

	ratesWorkerID = "disk_rates_sensors"
)

type sensors struct {
	totalReads  *diskIOSensor
	totalWrites *diskIOSensor
	readRate    *diskIOSensor
	writeRate   *diskIOSensor
}

// ioWorker creates sensors for disk IO counts and rates per device. It
// maintains an internal map of devices being tracked.
type ioWorker struct {
	logger  *slog.Logger
	devices map[string]*sensors
	mu      sync.Mutex
}

// addDevice adds a new device to the tracker map. If sthe device is already
// being tracked, it will not be added again. The bool return indicates whether
// a device was added (true) or not (false).
func (w *ioWorker) addDevice(dev *device) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.devices[dev.id]; !ok {
		w.devices[dev.id] = newDeviceSensors(dev)
	}
}

// updateDevice will update a tracked device's stats. For rates, it will
// recalculate based on the given time delta.
func (w *ioWorker) updateDevice(dev *device, stats map[stat]uint64, delta time.Duration) []sensor.Details {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.devices[dev.id].totalReads.update(stats, delta)
	w.devices[dev.id].totalWrites.update(stats, delta)
	w.devices[dev.id].readRate.update(stats, delta)
	w.devices[dev.id].writeRate.update(stats, delta)

	return []sensor.Details{
		w.devices[dev.id].totalReads,
		w.devices[dev.id].totalWrites,
		w.devices[dev.id].readRate,
		w.devices[dev.id].writeRate,
	}
}

func (w *ioWorker) updateTotals(stats map[stat]uint64, delta time.Duration) []sensor.Details {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.devices["total"].totalReads.update(stats, delta)
	w.devices["total"].totalWrites.update(stats, delta)
	w.devices["total"].readRate.update(stats, delta)
	w.devices["total"].writeRate.update(stats, delta)

	return []sensor.Details{
		w.devices["total"].totalReads,
		w.devices["total"].totalWrites,
		w.devices["total"].readRate,
		w.devices["total"].writeRate,
	}
}

func (w *ioWorker) Interval() time.Duration { return ratesUpdateInterval }

func (w *ioWorker) Jitter() time.Duration { return ratesUpdateJitter }

func (w *ioWorker) Sensors(_ context.Context, duration time.Duration) ([]sensor.Details, error) {
	// Get valid devices.
	deviceNames, err := getDeviceNames()
	if err != nil {
		return nil, fmt.Errorf("could not fetch disk devices: %w", err)
	}

	sensors := make([]sensor.Details, 0, 4*len(deviceNames)+4) //nolint:mnd
	totals := make(map[stat]uint64)

	// Get the current device info and stats for all valid devices.
	for _, name := range deviceNames {
		dev, stats, err := getDevice(name)
		if err != nil {
			w.logger.Warn("Unable to read device stats.", slog.Any("error", err))

			continue
		}

		// Add device (if it isn't already tracked).
		w.addDevice(dev)

		// Update device stats and return updated sensors.
		sensors = append(sensors, w.updateDevice(dev, stats, duration)...)

		// Don't include "aggregate" devices in totals.
		if strings.HasPrefix(dev.id, "dm") || strings.HasPrefix(dev.id, "md") {
			continue
		}
		// Add device stats to the totals.
		for stat, value := range stats {
			totals[stat] += value
		}
	}

	// Update total stats and return updated sensors.
	sensors = append(sensors, w.updateTotals(totals, duration)...)

	return sensors, nil
}

func NewIOWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	worker := &ioWorker{
		devices: make(map[string]*sensors),
		logger:  logging.FromContext(ctx).WithGroup(ratesWorkerID),
	}

	// Add sensors for a pseudo "total" device which tracks total values from
	// all devices.
	worker.devices["total"] = newDeviceSensors(&device{id: "total"})

	return &linux.SensorWorker{
			Value:    worker,
			WorkerID: ratesWorkerID,
		},
		nil
}

func newDeviceSensors(dev *device) *sensors {
	return &sensors{
		totalReads:  newDiskIOSensor(dev, linux.SensorDiskReads),
		totalWrites: newDiskIOSensor(dev, linux.SensorDiskWrites),
		readRate:    newDiskIORateSensor(dev, linux.SensorDiskReadRate),
		writeRate:   newDiskIORateSensor(dev, linux.SensorDiskWriteRate),
	}
}
