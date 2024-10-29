// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

// Worker is the base interface representing a worker that produces sensors or
// events. It has an ID and functions to start/stop producing sensors/events.
type Worker[T any] interface {
	ID() string
	Stop() error
	Start(ctx context.Context) (<-chan T, error)
}

// WorkerWithPreferences represents a worker that has preferences that can be
// set by a user.
type WorkerWithPreferences[T any, P any] interface {
	Worker[T]
	DefaultPreferences() P
}

// SensorWorker is a worker that produces sensors. In addition to the base
// worker methods, it has a function to generate a list of sensor values.
type SensorWorker interface {
	Worker[sensor.Entity]
	Sensors(ctx context.Context) ([]sensor.Entity, error)
}

// EventWorker is a worker that produces events. It does not extend further from
// the base worker other than defining the type of data produced.
type EventWorker interface {
	Worker[event.Event]
}

// startWorkers takes a slice of Workers of a particular type (sensor or event)
// and runs their start functions, logging any errors.
func startWorkers[T any](ctx context.Context, workers ...Worker[T]) []<-chan T {
	var eventCh []<-chan T

	for _, worker := range workers {
		logging.FromContext(ctx).Debug("Starting worker",
			slog.String("worker", worker.ID()))

		workerCh, err := worker.Start(ctx)
		if err != nil {
			logging.FromContext(ctx).
				Warn("Could not start worker.",
					slog.String("worker", worker.ID()),
					slog.Any("errors", err))
		} else {
			eventCh = append(eventCh, workerCh)
		}
	}

	return eventCh
}

// stopWorkers takes a slice of Workers of a particular type (sensor or event)
// and runs their stop functions, logging any errors.
func stopWorkers[T any](ctx context.Context, workers ...Worker[T]) {
	for _, worker := range workers {
		logging.FromContext(ctx).Debug("Stopping worker", slog.String("worker", worker.ID()))

		if err := worker.Stop(); err != nil {
			logging.FromContext(ctx).
				Warn("Could not stop worker.",
					slog.String("worker", worker.ID()),
					slog.Any("errors", err))
		}
	}
}

// processWorkers handles starting, stopping and processing data from a slice of
// workers passed in.  It will start the workers, monitor for data and send it
// to Home Assistant, and stop workers when the passed context is canceled.
func processWorkers[T any](ctx context.Context, hassclient *hass.Client, workers ...Worker[T]) {
	// Start all inactive workers of all controllers.
	workerOutputs := startWorkers(ctx, workers...)
	if len(workerOutputs) == 0 {
		logging.FromContext(ctx).Warn("No workers were started.")
		return
	}

	// When the context is done, stop all active workers of all controllers.
	go func() {
		<-ctx.Done()
		stopWorkers(ctx, workers...)
	}()

	// Process all events/sensors from all workers.
	for details := range mergeCh(ctx, workerOutputs...) {
		go func(e T) {
			var err error

			switch details := any(e).(type) {
			case sensor.Entity:
				err = hassclient.ProcessSensor(ctx, details)
			case event.Event:
				err = hassclient.ProcessEvent(ctx, details)
			}

			if err != nil {
				logging.FromContext(ctx).Error("Processing failed.", slog.Any("error", err))
			}
		}(details)
	}
}
