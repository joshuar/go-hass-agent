// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/linux/apps"
	"github.com/joshuar/go-hass-agent/internal/linux/battery"
	"github.com/joshuar/go-hass-agent/internal/linux/cpu"
	"github.com/joshuar/go-hass-agent/internal/linux/disk"
	"github.com/joshuar/go-hass-agent/internal/linux/location"
	"github.com/joshuar/go-hass-agent/internal/linux/mem"
	"github.com/joshuar/go-hass-agent/internal/linux/net"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/problems"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/linux/time"
	"github.com/joshuar/go-hass-agent/internal/linux/user"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

func newDevice(_ context.Context) *linux.Device {
	return linux.NewDevice(config.AppName, config.AppVersion)
}

// sensorWorkers returns a list of functions to start to enable sensor tracking.
func sensorWorkers() []func(context.Context) chan tracker.Sensor {
	var workers []func(context.Context) chan tracker.Sensor
	workers = append(workers,
		battery.Updater,
		apps.Updater,
		net.ConnectionsUpdater,
		net.RatesUpdater,
		problems.Updater,
		mem.Updater,
		cpu.LoadAvgUpdater,
		disk.UsageUpdater,
		time.Updater,
		power.ScreenLockUpdater,
		power.PowerStateUpdater,
		power.PowerProfileUpdater,
		user.Updater,
		system.Versions,
		// system.TempUpdater,
		system.HWSensorUpdater,
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
