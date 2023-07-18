// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

// SensorWorkers returns a list of functions to start to enable sensor tracking.
func SensorWorkers() []func(context.Context, chan interface{}) {
	var workers []func(context.Context, chan interface{})
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
	return workers
}

type LinuxDeviceAPI struct {
	dbus map[string]*linux.Bus
	mu   sync.Mutex
}

// EndPoint will return the given endpoint as an interface. Use
// device.GetAPIEndpoint to safely assert the type of the API.
func (d *LinuxDeviceAPI) EndPoint(e string) interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dbus[e]
}

// NewDeviceAPI sets up a DeviceAPI struct with appropriate DBus API endpoints.
func NewDeviceAPI(ctx context.Context) *LinuxDeviceAPI {
	api := new(LinuxDeviceAPI)
	api.dbus = make(map[string]*linux.Bus)
	api.mu.Lock()
	api.dbus["session"] = linux.NewBus(ctx, linux.SessionBus)
	api.dbus["system"] = linux.NewBus(ctx, linux.SystemBus)
	api.mu.Unlock()
	return api
}
