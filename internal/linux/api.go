// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"sync"
)

type LinuxDeviceAPI struct {
	mu   sync.Mutex
	dbus map[string]*bus
}

// NewDeviceAPI sets up a DeviceAPI struct with appropriate DBus API endpoints
func NewDeviceAPI(ctx context.Context) *LinuxDeviceAPI {
	api := new(LinuxDeviceAPI)
	api.dbus = make(map[string]*bus)
	api.mu.Lock()
	api.dbus["session"] = newBus(ctx, sessionBus)
	api.dbus["system"] = newBus(ctx, systemBus)
	api.mu.Unlock()
	return api
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
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dbus[e]
}
