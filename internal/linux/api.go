// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
)

type LinuxDeviceAPI struct {
	dbus map[string]*bus
}

// NewDeviceAPI sets up a DeviceAPI struct with appropriate DBus API endpoints
func NewDeviceAPI(ctx context.Context) *LinuxDeviceAPI {

	dbusEndpoints := make(map[string]*bus)
	dbusEndpoints["session"] = newBus(ctx, sessionBus)
	dbusEndpoints["system"] = newBus(ctx, systemBus)

	return &LinuxDeviceAPI{
		dbus: dbusEndpoints,
	}
}

func (d *LinuxDeviceAPI) SensorWorkers() []func(context.Context, chan interface{}) {
	var workers []func(context.Context, chan interface{})
	workers = append(workers, LocationUpdater)
	workers = append(workers, BatteryUpdater)
	workers = append(workers, AppUpdater)
	workers = append(workers, NetworkConnectionsUpdater)
	workers = append(workers, NetworkStatsUpdater)
	workers = append(workers, PowerUpater)
	workers = append(workers, ProblemsUpdater)
	workers = append(workers, MemoryUpdater)
	workers = append(workers, LoadAvgUpdater)
	workers = append(workers, DiskUsageUpdater)
	workers = append(workers, TimeUpdater)
	return workers
}

func (d *LinuxDeviceAPI) EndPoint(e string) interface{} {
	return d.dbus[e]
}
