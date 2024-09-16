// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	deviceControllerID = "device_controller"
)

type deviceController map[string]*workerState

func (w deviceController) ID() string {
	return deviceControllerID
}

func (w deviceController) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(w))

	for id, worker := range w {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (w deviceController) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(w))

	for id, worker := range w {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, id)
		}
	}

	return inactiveWorkers
}

func (w deviceController) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	worker, exists := w[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := w[name].Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	w[name].started = true

	return workerCh, nil
}

func (w deviceController) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := w[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}

func (w deviceController) States(ctx context.Context) []sensor.Details {
	var sensors []sensor.Details

	for _, worker := range w.ActiveWorkers() {
		workerSensors, err := w[worker].Sensors(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("controller", w.ID())).
				Debug("Could not retrieve worker sensors",
					slog.String("worker", w[worker].ID()),
					slog.Any("error", err))
		}

		sensors = append(sensors, workerSensors...)
	}

	return sensors
}

func (agent *Agent) newDeviceController(ctx context.Context, prefs *preferences.Preferences) SensorController {
	var worker worker

	controller := make(deviceController)

	// Set up sensor workers.
	worker = newVersionWorker(prefs.AgentVersion())
	controller[worker.ID()] = &workerState{worker: worker}
	worker = newExternalIPUpdaterWorker(ctx)
	controller[worker.ID()] = &workerState{worker: worker}
	worker = newConnectionLatencyWorker(prefs)
	controller[worker.ID()] = &workerState{worker: worker}

	return controller
}
