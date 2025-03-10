// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

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
	"github.com/joshuar/go-hass-agent/internal/models"
)

// sensorEventWorkersInitFuncs are all of the sensor workers that generate sensors on events.
var sensorEventWorkersInitFuncs = []func(ctx context.Context) (*linux.EventSensorWorker, error){
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

// sensorPollingWorkersInitFuncs are all of the sensor workers that need to poll to get their sensors.
var sensorPollingWorkersInitFuncs = []func(ctx context.Context) (*linux.PollingSensorWorker, error){
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

// sensorOneShotWorkersInitFuncs are all the sensor workers that run one-time to generate sensors.
var sensorOneShotWorkersInitFuncs = []func(ctx context.Context) (*linux.OneShotSensorWorker, error){
	system.NewCPUVulnerabilityWorker,
	system.NewInfoWorker,
	system.NewfwupdWorker,
	system.NewLastBootWorker,
}

// sensorLaptopWorkersInitFuncs are sensor workers that should only be run on laptops.
var sensorLaptopWorkersInitFuncs = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	power.NewLaptopWorker, location.NewLocationWorker,
}

// eventWorkersInitFuncs are event workers that produce events rather than sensors.
var eventWorkersInitFuncs = []func(ctx context.Context) (*linux.EventWorker, error){
	mem.NewOOMEventsWorker,
	system.NewUserSessionEventsWorker,
}

// setupOSWorkers creates slices of OS-specific sensor and event Workers that
// can be run by the agent. It handles initializing the workers with OS-specific
// data, reporting any errors.
func setupOSWorkers(ctx context.Context) []Worker[models.Entity] {
	workers := make([]Worker[models.Entity], 0,
		len(eventWorkersInitFuncs)+
			len(sensorPollingWorkersInitFuncs)+
			len(sensorEventWorkersInitFuncs)+
			len(sensorOneShotWorkersInitFuncs)+
			len(sensorLaptopWorkersInitFuncs))

	// Set up a logger.
	logger := logging.FromContext(ctx).With(slog.Group("linux", slog.String("controller", "sensor")))
	ctx = logging.ToContext(ctx, logger)

	for _, workerInit := range eventWorkersInitFuncs {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		workers = append(workers, worker)
	}

	for _, workerInit := range sensorEventWorkersInitFuncs {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		workers = append(workers, worker)
	}

	for _, workerInit := range sensorPollingWorkersInitFuncs {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		workers = append(workers, worker)
	}

	for _, workerInit := range sensorOneShotWorkersInitFuncs {
		worker, err := workerInit(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init worker.",
				slog.Any("error", err))

			continue
		}

		workers = append(workers, worker)
	}

	// Get the type of device we are running on.
	chassis, _ := info.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	laptops := []string{"Portable", "Laptop", "Notebook"}
	// If running on a laptop chassis, add laptop specific sensor workers.
	if slices.Contains(laptops, chassis) {
		for _, workerInit := range sensorLaptopWorkersInitFuncs {
			worker, err := workerInit(ctx)
			if err != nil {
				logging.FromContext(ctx).Warn("Could not init worker.",
					slog.Any("error", err))

				continue
			}

			workers = append(workers, worker)
		}
	}

	return workers
}
