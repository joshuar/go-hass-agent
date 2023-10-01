// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func newDevice(ctx context.Context) *linux.LinuxDevice {
	return linux.NewDevice(ctx, name, version)
}

// sensorWorkers returns a list of functions to start to enable sensor tracking.
func sensorWorkers() []func(context.Context, device.SensorTracker) {
	var workers []func(context.Context, device.SensorTracker)
	workers = append(workers,
		linux.LocationUpdater,
		linux.BatteryUpdater,
		linux.AppUpdater,
		linux.NetworkConnectionsUpdater,
		linux.NetworkStatsUpdater,
		linux.PowerUpater,
		linux.ProblemsUpdater,
		linux.MemoryUpdater,
		linux.LoadAvgUpdater,
		linux.DiskUsageUpdater,
		linux.TimeUpdater,
		linux.ScreenLockUpdater,
		linux.UsersUpdater,
		linux.Versions)
	return workers
}
