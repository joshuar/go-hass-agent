// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type deviceController struct {
	sensorWorkers map[string]*sensorWorker
	logger        *slog.Logger
}

func (w *deviceController) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(w.sensorWorkers))

	for id, worker := range w.sensorWorkers {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (w *deviceController) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(w.sensorWorkers))

	for id, worker := range w.sensorWorkers {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, id)
		}
	}

	return inactiveWorkers
}

func (w *deviceController) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	worker, exists := w.sensorWorkers[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := w.sensorWorkers[name].object.Updates(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	w.sensorWorkers[name].started = true

	return workerCh, nil
}

func (w *deviceController) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := w.sensorWorkers[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.object.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}

func (w *deviceController) StartAll(ctx context.Context) (<-chan sensor.Details, error) {
	outCh := make([]<-chan sensor.Details, 0, len(allworkers))

	var errs error

	for id := range w.sensorWorkers {
		workerCh, err := w.Start(ctx, id)
		if err != nil {
			errs = errors.Join(errs, err)

			continue
		}

		outCh = append(outCh, workerCh)
	}

	return sensor.MergeSensorCh(ctx, outCh...), errs
}

func (w *deviceController) StopAll() error {
	var errs error

	for id := range w.sensorWorkers {
		if err := w.Stop(id); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func newDeviceController(ctx context.Context) SensorController {
	var worker Worker

	controller := &deviceController{
		sensorWorkers: make(map[string]*sensorWorker),
		logger:        logging.FromContext(ctx).With(slog.Group("device")),
	}

	// Set up sensor workers.
	worker = device.NewVersionWorker()
	controller.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}
	worker = device.NewExternalIPUpdaterWorker(ctx)
	controller.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}

	return controller
}
