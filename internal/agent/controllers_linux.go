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

// sensorEventWorkers are all of the sensor workers that generate sensors on events.
var sensorEventWorkers = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	apps.NewAppWorker,
	battery.NewBatteryWorker,
	net.NewConnectionWorker,
	net.NewAddressWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	desktop.NewDesktopWorker,
}

// sensorPollingWorkers are all of the sensor workers that need to poll to get their sensors.
var sensorPollingWorkers = []func(ctx context.Context) (*linux.PollingSensorWorker, error){
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

// sensorOneShotWorkers are all the sensor workers that run one-time to generate sensors.
var sensorOneShotWorkers = []func(ctx context.Context) (*linux.OneShotSensorWorker, error){
	cpu.NewCPUVulnerabilityWorker,
	system.NewInfoWorker,
	system.NewfwupdWorker,
}

// sensorLaptopWorkers are sensor workers that should only be run on laptops.
var sensorLaptopWorkers = []func(ctx context.Context) (*linux.EventSensorWorker, error){
	power.NewLaptopWorker, location.NewLocationWorker,
}

var eventWorkers = []func(ctx context.Context) (*linux.EventWorker, error){}

const (
	linuxSensorControllerID = "linux_sensors_controller"
	linuxEventControllerID  = "linux_events_controller"
)

var (
	ErrWorkerAlreadyStarted = errors.New("worker already started")
	ErrUnknownWorker        = errors.New("unknown worker")
)

// newOperatingSystemControllers initializes the list of sensor workers for sensors and
// returns those that are supported on this device.
func newOperatingSystemControllers(ctx context.Context) (SensorController, EventController) {
	// Set up the context.
	ctx = linux.NewContext(ctx)
	// Set up a logger.
	logger := logging.FromContext(ctx).With(slog.Group("linux", slog.String("controller", "sensor")))
	ctx = logging.ToContext(ctx, logger)
	// Create the controllers.
	return newOperatingSystemSensorController(ctx), newOperatingSystemEventController(ctx)
}

func newOperatingSystemSensorController(ctx context.Context) SensorController {
	sensorController := &sensorController{
		id:      linuxSensorControllerID,
		workers: make(map[string]*sensorWorkerState),
	}

	for _, startWorker := range sensorEventWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logging.FromContext(ctx).Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			sensorController.workers[worker.ID()] = &sensorWorkerState{sensorWorker: worker}
		}
	}

	for _, startWorker := range sensorPollingWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logging.FromContext(ctx).Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			sensorController.workers[worker.ID()] = &sensorWorkerState{sensorWorker: worker}
		}
	}

	for _, startWorker := range sensorOneShotWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logging.FromContext(ctx).Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			sensorController.workers[worker.ID()] = &sensorWorkerState{sensorWorker: worker}
		}
	}

	// Get the type of device we are running on.
	chassis, _ := device.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	// If running on a laptop, add laptop specific sensor workers.
	if chassis == "laptop" {
		for _, startWorker := range sensorLaptopWorkers {
			if worker, err := startWorker(ctx); err != nil {
				logging.FromContext(ctx).Warn("Could not add worker.",
					slog.String("worker", worker.ID()),
					slog.Any("error", err))
			} else {
				sensorController.workers[worker.ID()] = &sensorWorkerState{sensorWorker: worker}
			}
		}
	}

	return sensorController
}

func newOperatingSystemEventController(ctx context.Context) EventController {
	// Create an events controller.
	eventController := &eventController{
		id:      linuxEventControllerID,
		workers: make(map[string]*eventWorkerState),
	}

	for _, startWorker := range eventWorkers {
		if worker, err := startWorker(ctx); err != nil {
			logging.FromContext(ctx).Warn("Could not add worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))
		} else {
			eventController.workers[worker.ID()] = &eventWorkerState{eventWorker: worker}
		}
	}

	return eventController
}
