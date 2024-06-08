// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/rs/zerolog/log"

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

// workers is the list of sensor workers supported on Linux.
var workers = []func() (*linux.SensorWorker, error){
	apps.NewAppWorker,
	battery.NewBatteryWorker,
	cpu.NewLoadAvgWorker,
	cpu.NewUsageWorker,
	desktop.NewDesktopWorker,
	disk.NewIOWorker,
	disk.NewUsageWorker,
	location.NewLocationWorker,
	mem.NewUsageWorker,
	net.NewConnectionWorker,
	net.NewRatesWorker,
	power.NewLaptopWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	problems.NewProblemsWorker,
	// power.IdleUpdater,
	system.NewHWMonWorker,
	system.NewInfoWorker,
	system.NewTimeWorker,
	user.NewUserWorker,
}

func newDevice(_ context.Context) *linux.Device {
	return linux.NewDevice(preferences.AppName, preferences.AppVersion)
}

// sensorWorkers initialises the list of workers for sensors and returns those
// that are supported on this device.
func sensorWorkers() []Worker {
	activeWorkers := make([]Worker, 0, len(workers))

	for _, w := range workers {
		worker, err := w()
		if err != nil {
			log.Warn().Err(err).Msg("Could not activate a worker.")

			continue
		}

		activeWorkers = append(activeWorkers, worker)
	}
	return activeWorkers
}

// setupDeviceContext returns a new Context that contains the D-Bus API.
func setupDeviceContext(ctx context.Context) context.Context {
	return dbusx.Setup(ctx)
}
