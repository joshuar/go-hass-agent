// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/linux/apps"
	"github.com/joshuar/go-hass-agent/internal/linux/battery"
	"github.com/joshuar/go-hass-agent/internal/linux/cpu"
	"github.com/joshuar/go-hass-agent/internal/linux/desktop"
	"github.com/joshuar/go-hass-agent/internal/linux/disk"
	"github.com/joshuar/go-hass-agent/internal/linux/location"
	"github.com/joshuar/go-hass-agent/internal/linux/mem"
	"github.com/joshuar/go-hass-agent/internal/linux/net"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/problems"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/linux/user"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type Controller interface {
	Sensors() []sensor.Details
}

type Polling interface {
	Controller
	Interval() time.Duration
	Jitter() time.Duration
}

type Events interface {
	Controller
	Events(ctx context.Context) chan sensor.Details
}

func newDevice(_ context.Context) *linux.Device {
	return linux.NewDevice(preferences.AppName, preferences.AppVersion)
}

// sensorWorkers returns a list of functions to start to enable sensor tracking.
func sensorWorkers() []func(context.Context) chan sensor.Details {
	var workers []func(context.Context) chan sensor.Details
	workers = append(workers,
		battery.Updater,
		apps.Updater,
		net.ConnectionsUpdater,
		net.RatesUpdater,
		problems.Updater,
		mem.Updater,
		cpu.LoadAvgUpdater,
		cpu.UsageUpdater,
		disk.UsageUpdater,
		disk.IOUpdater,
		power.ScreenLockUpdater,
		power.LaptopUpdater,
		power.StateUpdater,
		power.ProfileUpdater,
		power.IdleUpdater,
		user.Updater,
		system.Versions,
		system.HWSensorUpdater,
		system.UptimeUpdater,
		desktop.Updater,
	)
	return workers
}

func locationWorker() func(context.Context) chan *hass.LocationData {
	return location.Updater
}

// Setup returns a new Context that contains the D-Bus API.
func setupDeviceContext(ctx context.Context) context.Context {
	return dbusx.Setup(ctx)
}
