// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"
	"log/slog"
	"slices"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/device/info"
	"github.com/joshuar/go-hass-agent/internal/linux/battery"
	"github.com/joshuar/go-hass-agent/internal/linux/cpu"
	"github.com/joshuar/go-hass-agent/internal/linux/desktop"
	"github.com/joshuar/go-hass-agent/internal/linux/disk"
	"github.com/joshuar/go-hass-agent/internal/linux/location"
	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/mem"
	"github.com/joshuar/go-hass-agent/internal/linux/net"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

var linuxWorkers = []func(ctx context.Context) (workers.EntityWorker, error){
	battery.NewBatteryWorker,
	disk.NewIOWorker,
	disk.NewUsageWorker,
	disk.NewSmartWorker,
	cpu.NewUsageWorker,
	cpu.NewLoadAvgWorker,
	cpu.NewFreqWorker,
	desktop.NewAppStateWorker,
	desktop.NewDesktopWorker,
	media.NewMicUsageWorker,
	media.NewWebcamUsageWorker,
	mem.NewUsageWorker,
	mem.NewOOMEventsWorker,
	net.NewConnectionWorker,
	net.NewAddressWorker,
	net.NewNetStatsWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	system.NewChronyWorker,
	system.NewfwupdWorker,
	system.NewHWMonWorker,
	system.NewInfoWorker,
	system.NewLastBootWorker,
	system.NewProblemsWorker,
	system.NewUptimeTimeWorker,
	system.NewUserSessionEventsWorker,
	system.NewCPUVulnerabilityWorker,
}

// linuxLaptopWorkers are sensor workers that should only be run on laptops.
var linuxLaptopWorkers = []func(ctx context.Context) (workers.EntityWorker, error){
	power.NewLaptopWorker, location.NewLocationWorker,
}

// CreateOSEntityWorkers sets up all OS-specific entity workers.
func CreateOSEntityWorkers(ctx context.Context) []workers.EntityWorker {
	osWorkers := make([]workers.EntityWorker, 0, len(linuxWorkers)+len(linuxLaptopWorkers))

	for workerInit := range slices.Values(linuxWorkers) {
		worker, err := workerInit(ctx)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		osWorkers = append(osWorkers, worker)
	}

	// Get the type of device we are running on.
	chassis, _ := info.Chassis()
	laptops := []string{"Portable", "Laptop", "Notebook"}
	// If running on a laptop chassis, add laptop specific sensor workers.
	if slices.Contains(laptops, chassis) {
		for _, workerInit := range linuxLaptopWorkers {
			worker, err := workerInit(ctx)
			if err != nil {
				slogctx.FromCtx(ctx).Warn("Could not init worker.",
					slog.Any("error", err))

				continue
			}

			osWorkers = append(osWorkers, worker)
		}
	}

	return osWorkers
}
