// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/device"
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
	"github.com/joshuar/go-hass-agent/internal/logging"
)

// eventWorkers are all of the sensor workers that generate sensors on events.
var eventWorkers = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	apps.NewAppWorker,
	battery.NewBatteryWorker,
	net.NewConnectionWorker,
	net.NewAddressWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	desktop.NewDesktopWorker,
}

// pollingWorkers are all of the sensor workers that need to poll to get their sensors.
var pollingWorkers = []func(ctx context.Context) (*linux.PollingSensorWorker, error){
	disk.NewIOWorker,
	disk.NewUsageWorker,
	cpu.NewLoadAvgWorker,
	cpu.NewUsageWorker,
	mem.NewUsageWorker,
	net.NewNetStatsWorker,
	problems.NewProblemsWorker,
	system.NewTimeWorker,
	system.NewHWMonWorker,
}

// oneShotWorkers are all the sensor workers that run one-time to generate sensors.
var oneShotWorkers = []func(ctx context.Context) (*linux.OneShotSensorWorker, error){
	system.NewInfoWorker,
	system.NewfwupdWorker,
}

// laptopWorkers are sensor workers that should only be run on laptops.
var laptopWorkers = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	power.NewLaptopWorker, location.NewLocationWorker,
}

const (
	linuxSensorControllerID = "linux_sensors"
)

var (
	ErrWorkerAlreadyStarted = errors.New("worker already started")
	ErrUnknownWorker        = errors.New("unknown worker")
)

// newOSSensorController initializes the list of sensor workers for sensors and
// returns those that are supported on this device.
func newOSSensorController(ctx context.Context) SensorController {
	ctx = linux.NewContext(ctx)

	logger := logging.FromContext(ctx).With(slog.Group("linux", slog.String("controller", "sensor")))
	ctx = logging.ToContext(ctx, logger)
	controller := &sensorController{
		id:      linuxSensorControllerID,
		workers: make(map[string]*workerState),
	}

	for _, startWorker := range eventWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logger.Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			controller.workers[worker.ID()] = &workerState{worker: worker}
		}
	}

	for _, startWorker := range pollingWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logger.Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			controller.workers[worker.ID()] = &workerState{worker: worker}
		}
	}

	for _, startWorker := range oneShotWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logger.Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			controller.workers[worker.ID()] = &workerState{worker: worker}
		}
	}

	// Get the type of device we are running on.
	chassis, _ := device.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	// If running on a laptop, add laptop specific sensor workers.
	if chassis == "laptop" {
		for _, startWorker := range laptopWorkers {
			if worker, err := startWorker(ctx); err != nil {
				logger.Warn("Could not add worker.",
					slog.String("worker", worker.ID()),
					slog.Any("error", err))
			} else {
				controller.workers[worker.ID()] = &workerState{worker: worker}
			}
		}
	}

	return controller
}
