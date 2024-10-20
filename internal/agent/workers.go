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

type Worker[T any] interface {
	ID() string
	Stop() error
	Start(ctx context.Context) (<-chan T, error)
}

type SensorWorker interface {
	Worker[sensor.Entity]
	Sensors(ctx context.Context) ([]sensor.Entity, error)
}

type EventWorker interface {
	Worker[event.Event]
}

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
