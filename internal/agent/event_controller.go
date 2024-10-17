// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

// runEventWorkers will start all the sensor worker functions for all event
// controllers passed in. It returns a single merged channel of events.
//
//nolint:gocognit
func runEventWorkers(ctx context.Context, prefs *preferences.Preferences, controllers ...EventController) {
	var eventCh []<-chan event.Event

	for _, controller := range controllers {
		logging.FromContext(ctx).Debug("Running controller", slog.String("controller", controller.ID()))

		for _, workerName := range controller.InactiveWorkers() {
			logging.FromContext(ctx).Debug("Starting worker", slog.String("worker", workerName))

			workerCh, err := controller.Start(ctx, workerName)
			if err != nil {
				logging.FromContext(ctx).
					Warn("Could not start worker.",
						slog.String("controller", controller.ID()),
						slog.String("worker", workerName),
						slog.Any("errors", err))
			} else {
				eventCh = append(eventCh, workerCh)
			}
		}
	}

	if len(eventCh) == 0 {
		logging.FromContext(ctx).Warn("No workers were started by any controllers.")
		return
	}

	hassclient, err := hass.NewClient(ctx)
	if err != nil {
		logging.FromContext(ctx).Debug("Cannot create Home Assistant client.", slog.Any("error", err))
		return
	}

	hassclient.Endpoint(prefs.RestAPIURL(), hass.DefaultTimeout)

	go func() {
		<-ctx.Done()
		logging.FromContext(ctx).Debug("Stopping all event controllers.")

		for _, controller := range controllers {
			for _, workerName := range controller.ActiveWorkers() {
				if err := controller.Stop(workerName); err != nil {
					logging.FromContext(ctx).
						Warn("Could not stop worker.",
							slog.String("controller", controller.ID()),
							slog.String("worker", workerName),
							slog.Any("errors", err))
				}
			}
		}
	}()

	logging.FromContext(ctx).Debug("Processing sensor updates.")

	for details := range mergeCh(ctx, eventCh...) {
		go func(details event.Event) {
			if err := hassclient.ProcessEvent(ctx, details); err != nil {
				logging.FromContext(ctx).Error("Process sensor failed.", slog.Any("error", err))
			}
		}(details)
	}
}
