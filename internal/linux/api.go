// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
)

type DeviceAPI struct {
	dBusSystem, dBusSession *bus
	dbus                    map[string]*bus
	Workers                 []func(context.Context, chan interface{})
}

// NewDeviceAPI sets up a DeviceAPI struct with appropriate DBus API endpoints
func NewDeviceAPI(ctx context.Context) *DeviceAPI {
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

	dbusEndpoints := make(map[string]*bus)
	dbusEndpoints["session"] = newBus(ctx, sessionBus)
	dbusEndpoints["system"] = newBus(ctx, systemBus)

	api := &DeviceAPI{
		dBusSystem:  dbusEndpoints["system"],
		dBusSession: dbusEndpoints["session"],
		dbus:        dbusEndpoints,	
		Workers:     workers,
	}
	if api.dBusSystem == nil && api.dBusSession == nil {
		return nil
	} else {
		// go api.monitorDBus(ctx)
		return api
	}
}

func (d *DeviceAPI) SensorWorkers() []func(context.Context, chan interface{}) {
	return d.Workers
}

func (d *DeviceAPI) EndPoint(e string) interface{} {
	return d.dbus[e]
}

// SessionBusRequest creates a request builder for the session bus
func (d *DeviceAPI) SessionBusRequest() *busRequest {
	return &busRequest{
		bus: d.dBusSession,
	}
}

// SystemBusRequest creates a request builder for the system bus
func (d *DeviceAPI) SystemBusRequest() *busRequest {
	return &busRequest{
		bus: d.dBusSystem,
	}
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for DeviceAPI values in Contexts. It is
// unexported; clients use linux.NewContext and linux.FromContext
// instead of using this key directly.
var configKey key

// StoreAPIInContext returns a new Context that embeds a DeviceAPI.
func StoreAPIInContext(ctx context.Context, c *DeviceAPI) context.Context {
	return context.WithValue(ctx, configKey, c)
}

// FetchAPIFromContext returns the DeviceAPI stored in ctx, or an error if there
// is none
func FetchAPIFromContext(ctx context.Context) (*DeviceAPI, error) {
	if c, ok := ctx.Value(configKey).(*DeviceAPI); !ok {
		return nil, errors.New("no API in context")
	} else {
		return c, nil
	}
}
