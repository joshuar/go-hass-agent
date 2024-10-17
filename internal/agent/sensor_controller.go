// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type sensorWorker interface {
	ID() string
	Start(ctx context.Context) (<-chan sensor.Entity, error)
	Stop() error
	Sensors(ctx context.Context) ([]sensor.Entity, error)
}

type sensorWorkerState struct {
	sensorWorker
	started bool
}

type sensorController struct {
	workers map[string]*sensorWorkerState
	id      string
}

func (c *sensorController) ID() string {
	return c.id
}

func (c *sensorController) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(c.workers))

	for id, worker := range c.workers {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (c *sensorController) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(c.workers))

	for id, worker := range c.workers {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, id)
		}
	}

	return inactiveWorkers
}

func (c *sensorController) Start(ctx context.Context, name string) (<-chan sensor.Entity, error) {
	worker, exists := c.workers[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := c.workers[name].Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	c.workers[name].started = true

	return workerCh, nil
}

func (c *sensorController) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := c.workers[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}

func (c *sensorController) States(ctx context.Context) []sensor.Entity {
	var sensors []sensor.Entity

	for _, workerID := range c.ActiveWorkers() {
		worker, found := c.workers[workerID]
		if !found {
			logging.FromContext(ctx).
				With(slog.String("controller", c.ID())).
				Debug("Worker not found",
					slog.String("worker", workerID))

			continue
		}

		workerSensors, err := worker.Sensors(ctx)
		if err != nil || len(workerSensors) == 0 {
			logging.FromContext(ctx).
				With(slog.String("controller", c.ID())).
				Debug("Could not retrieve worker sensors",
					slog.String("worker", worker.ID()),
					slog.Any("error", err))

			continue
		}

		sensors = append(sensors, workerSensors...)
	}

	return sensors
}
