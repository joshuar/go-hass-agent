// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/event"
)

type eventWorker interface {
	ID() string
	Start(ctx context.Context) (<-chan event.Event, error)
	Stop() error
}

type eventWorkerState struct {
	eventWorker
	started bool
}

type eventController struct {
	workers map[string]*eventWorkerState
	id      string
}

func (e *eventController) ID() string {
	return e.id
}

func (e *eventController) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(e.workers))

	for id, worker := range e.workers {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (e *eventController) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(e.workers))

	for id, worker := range e.workers {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, id)
		}
	}

	return inactiveWorkers
}

func (e *eventController) Start(ctx context.Context, name string) (<-chan event.Event, error) {
	worker, exists := e.workers[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := e.workers[name].Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	e.workers[name].started = true

	return workerCh, nil
}

func (e *eventController) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := e.workers[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}
