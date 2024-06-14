// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

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

// allworkers is the list of sensor allworkers supported on Linux.
var allworkers = []func() (*linux.SensorWorker, error){
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
	control context.CancelFunc
}

type linuxWorkers map[string]*workerControl

func (w linuxWorkers) ActiveWorkers() []string {
	workers := make([]string, 0, len(w))

	for _, worker := range w {
		if worker.control != nil {
			workers = append(workers, worker.object.Name())
		}
	}

	return workers
}

func (w linuxWorkers) InactiveWorkers() []string {
	workers := make([]string, 0, len(w))

	for _, worker := range w {
		if worker.control == nil {
			workers = append(workers, worker.object.Name())
		}
	}

	return workers
}

func (w linuxWorkers) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	if worker, ok := w[name]; ok {
		if worker.control != nil {
			return nil, ErrWorkerAlreadyStarted
		}

		workerCtx, workerCancelFunc := context.WithCancel(ctx)

		workerCh, err := w[name].object.Updates(workerCtx)
		if err != nil {
			return nil, fmt.Errorf("could not start worker: %w", err)
		}

		w[name].control = workerCancelFunc

		return workerCh, nil
	}

	return nil, ErrUnknownWorker
}

func (w linuxWorkers) Stop(name string) error {
	var worker *workerControl

	var exists bool

	if worker, exists = w[name]; !exists {
		return ErrUnknownWorker
	}

	worker.control()

	return nil
}

func (w linuxWorkers) StartAll(ctx context.Context) (<-chan sensor.Details, error) {
	outCh := make([]<-chan sensor.Details, 0, len(allworkers))

	var allerr error

	log.Debug().Msg("Starting all Linux workers.")

	for name, worker := range w {
		workerCtx, cancelFunc := context.WithCancel(ctx)

		log.Debug().Str("name", name).Str("description", worker.object.Description()).Msg("Starting sensor worker.")

		workerCh, err := worker.object.Updates(workerCtx)
		if err != nil {
			allerr = errors.Join(allerr, err)

			continue
		}

		outCh = append(outCh, workerCh)
		worker.control = cancelFunc
	}

	return sensor.MergeSensorCh(ctx, outCh...), allerr
}

func (w linuxWorkers) StopAll() error {
	for _, worker := range w {
		worker.control()
	}

	return nil
}

func newDevice(_ context.Context) *linux.Device {
	return linux.NewDevice(preferences.AppName, preferences.AppVersion)
}

// createSensorWorkers initialises the list of workers for sensors and returns those
// that are supported on this device.
//
//nolint:exhaustruct
func createSensorWorkers() WorkerController {
	workers := make(linuxWorkers)

	for _, w := range allworkers {
		worker, err := w()
		if err != nil {
			log.Debug().Err(err).Msg("Could not initialise worker.")

			continue
		}

		workers[worker.Name()] = &workerControl{object: worker}
	}

	return workers
}

// setupDeviceContext returns a new Context that contains the D-Bus API.
func setupDeviceContext(ctx context.Context) context.Context {
	return dbusx.Setup(ctx)
}
