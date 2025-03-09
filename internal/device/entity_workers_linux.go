// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"
	"log/slog"
	"slices"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/device/info"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

// linuxSensorEventWorkers are all of the sensor workers that generate sensors on events.
var linuxSensorEventWorkers = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	battery.NewBatteryWorker,
	media.NewMicUsageWorker,
	media.NewWebcamUsageWorker,
	net.NewConnectionWorker,
	net.NewAddressWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	desktop.NewAppWorker,
	desktop.NewDesktopWorker,
}

// eventWorkersInitFuncs are event workers that produce events rather than sensors.
var linuxEventWorkers = []func(ctx context.Context) (*linux.EventWorker, error){
	mem.NewOOMEventsWorker,
	system.NewUserSessionEventsWorker,
}

// linuxPollingWorkers are all of the sensor workers that need to poll to get their sensors.
var linuxPollingWorkers = []func(ctx context.Context) (*linux.PollingSensorWorker, error){
	disk.NewIOWorker,
	disk.NewUsageWorker,
	cpu.NewLoadAvgWorker,
	cpu.NewUsageWorker,
	cpu.NewFreqWorker,
	mem.NewUsageWorker,
	net.NewNetStatsWorker,
	system.NewProblemsWorker,
	system.NewUptimeTimeWorker,
	system.NewChronyWorker,
	system.NewHWMonWorker,
}

// linuxOneShotWorkers are all the sensor workers that run one-time to generate sensors.
var linuxOneShotWorkers = []func(ctx context.Context) (*linux.OneShotSensorWorker, error){
	system.NewCPUVulnerabilityWorker,
	system.NewInfoWorker,
	system.NewfwupdWorker,
	system.NewLastBootWorker,
}

// linuxLaptopWorkers are sensor workers that should only be run on laptops.
var linuxLaptopWorkers = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	power.NewLaptopWorker, location.NewLocationWorker,
}

// CreateOSEntityWorkers sets up all OS-specific entity workers.
func CreateOSEntityWorkers(ctx context.Context) []workers.EntityWorker {
	osWorkers := make([]workers.EntityWorker, 0,
		len(linuxSensorEventWorkers)+
			len(linuxPollingWorkers)+
			len(linuxSensorEventWorkers)+
			len(linuxOneShotWorkers)+
			len(linuxLaptopWorkers))

	// Set up a logger.
	logger := logging.FromContext(ctx).With(slog.Group("linux", slog.String("controller", "sensor")))
	ctx = logging.ToContext(ctx, logger)

	for _, workerInit := range linuxEventWorkers {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		osWorkers = append(osWorkers, worker)
	}

	for _, workerInit := range linuxSensorEventWorkers {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		osWorkers = append(osWorkers, worker)
	}

	for _, workerInit := range linuxPollingWorkers {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		osWorkers = append(osWorkers, worker)
	}

	for _, workerInit := range linuxOneShotWorkers {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		osWorkers = append(osWorkers, worker)
	}

	// Get the type of device we are running on.
	chassis, _ := info.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	laptops := []string{"Portable", "Laptop", "Notebook"}
	// If running on a laptop chassis, add laptop specific sensor workers.
	if slices.Contains(laptops, chassis) {
		for _, workerInit := range linuxLaptopWorkers {
			worker, err := workerInit(ctx)
			if err != nil {
				logging.FromContext(ctx).Warn("Could not init worker.",
					slog.Any("error", err))

				continue
			}

			osWorkers = append(osWorkers, worker)
		}
	}

	return osWorkers
}
