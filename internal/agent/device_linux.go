// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

func newDevice(_ context.Context) *linux.Device {
	return linux.NewDevice(config.AppName, config.AppVersion)
}

// sensorWorkers returns a list of functions to start to enable sensor tracking.
func sensorWorkers() []func(context.Context) chan tracker.Sensor {
	var workers []func(context.Context) chan tracker.Sensor
	workers = append(workers,
		linux.BatteryUpdater,
		linux.AppUpdater,
		linux.NetworkConnectionsUpdater,
		linux.NetworkStatsUpdater,
		linux.PowerProfileUpater,
		linux.ProblemsUpdater,
		linux.MemoryUpdater,
		linux.LoadAvgUpdater,
		linux.DiskUsageUpdater,
		linux.TimeUpdater,
		linux.ScreenLockUpdater,
		linux.UsersUpdater,
		linux.Versions,
		linux.TempUpdater,
		linux.PowerStateUpdater)
	return workers
}

func locationWorker() func(context.Context) chan tracker.Location {
	return linux.LocationUpdater
}
