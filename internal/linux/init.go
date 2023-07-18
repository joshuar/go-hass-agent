// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"sync"
)

var dbusAPI *LinuxDeviceAPI

type LinuxDeviceAPI struct {
	dbus map[dbusType]*Bus
	mu   sync.Mutex
}

// EndPoint will return the given endpoint as an interface. Use
// device.GetAPIEndpoint to safely assert the type of the API.
func (d *LinuxDeviceAPI) EndPoint(e dbusType) *Bus {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dbus[e]
}

// NewDeviceAPI sets up a DeviceAPI struct with appropriate DBus API endpoints.
func init() {
	ctx := context.Background()
	api := new(LinuxDeviceAPI)
	api.dbus = make(map[dbusType]*Bus)
	api.mu.Lock()
	api.dbus[SessionBus] = NewBus(ctx, SessionBus)
	api.dbus[SystemBus] = NewBus(ctx, SystemBus)
	api.mu.Unlock()
	dbusAPI = api
}
