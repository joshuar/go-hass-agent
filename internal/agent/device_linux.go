// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"fmt"

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
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

// allworkers is the list of sensor allworkers supported on Linux.
var allworkers = []func(context.Context) (*linux.SensorWorker, error){
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

var (
	ErrWorkerAlreadyStarted = errors.New("worker already started")
	ErrUnknownWorker        = errors.New("unknown worker")
)

type workerControl struct {
	object  Worker
	started bool
}

type linuxWorkers map[string]*workerControl

func (w linuxWorkers) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(w))

	for id, worker := range w {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (w linuxWorkers) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(w))

	for _, worker := range w {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, worker.object.ID())
		}
	}

	return inactiveWorkers
}

func (w linuxWorkers) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	worker, exists := w[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := w[name].object.Updates(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	w[name].started = true

	return workerCh, nil
}

func (w linuxWorkers) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := w[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.object.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}

func (w linuxWorkers) StartAll(ctx context.Context) (<-chan sensor.Details, error) {
	outCh := make([]<-chan sensor.Details, 0, len(allworkers))

	var errs error

	for id := range w {
		workerCh, err := w.Start(ctx, id)
		if err != nil {
			errs = errors.Join(errs, err)

			continue
		}

		outCh = append(outCh, workerCh)
	}

	return sensor.MergeSensorCh(ctx, outCh...), errs
}

func (w linuxWorkers) StopAll() error {
	var errs error

	for id := range w {
		if err := w.Stop(id); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// createSensorWorkers initialises the list of workers for sensors and returns those
// that are supported on this device.
//
//nolint:exhaustruct
func createSensorWorkers(ctx context.Context) WorkerController {
	workers := make(linuxWorkers)

	for _, startWorkerFunc := range allworkers {
		worker, err := startWorkerFunc(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not start a sensor worker.", "error", err.Error())

			continue
		}

		workers[worker.ID()] = &workerControl{object: worker}
	}

	return workers
}

// setupDeviceContext returns a new Context that contains the D-Bus API.
func setupDeviceContext(ctx context.Context) context.Context {
	return dbusx.Setup(ctx, logging.FromContext(ctx))
}
