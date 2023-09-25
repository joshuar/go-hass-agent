// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

// SensorWorkers returns a list of functions to start to enable sensor tracking.
func SensorWorkers() []func(context.Context, device.SensorTracker) {
	var workers []func(context.Context, device.SensorTracker)
	workers = append(workers, linux.LocationUpdater)
	workers = append(workers, linux.BatteryUpdater)
	workers = append(workers, linux.AppUpdater)
	workers = append(workers, linux.NetworkConnectionsUpdater)
	workers = append(workers, linux.NetworkStatsUpdater)
	workers = append(workers, linux.PowerUpater)
	workers = append(workers, linux.ProblemsUpdater)
	workers = append(workers, linux.MemoryUpdater)
	workers = append(workers, linux.LoadAvgUpdater)
	workers = append(workers, linux.DiskUsageUpdater)
	workers = append(workers, linux.TimeUpdater)
	workers = append(workers, linux.ScreenLockUpdater)
	workers = append(workers, linux.UsersUpdater)
	workers = append(workers, linux.Versions)
	return workers
}
