// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run github.com/matryer/moq -out worker_mocks_test.go . PollingSensorType EventSensorType OneShotSensorType EventType
package linux

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrUnknownWorker = errors.New("unknown sensor worker type")

// PollingSensorType interface represents sensors that are generated on some poll interval.
type PollingSensorType interface {
	UpdateDelta(delta time.Duration)
	Sensors(ctx context.Context) ([]models.Entity, error)
}

// EventSensorType interface represents sensors that are generated on some event
// trigger, such as D-Bus messages.
type EventSensorType interface {
	Events(ctx context.Context) (<-chan models.Entity, error)
	Sensors(ctx context.Context) ([]models.Entity, error)
}

// OneShotSensorType interface represents sensors that are generated only one-time and
// have no ongoing updates.
type OneShotSensorType interface {
	Sensors(ctx context.Context) ([]models.Entity, error)
}

type EventType interface {
	Events(ctx context.Context) (<-chan models.Entity, error)
}

// Worker is a struct embedded in other specific worker structs. It holds
// a function to cancel the worker and its ID.
type Worker struct {
	cancelFunc context.CancelFunc
	WorkerID   string
}

// ID is a name that can be used as an ID to represent the group of sensors
// managed by this worker.
func (w *Worker) ID() string {
	if w != nil {
		return w.WorkerID
	}

	return "Unknown Worker"
}

// Stop will stop any processing of sensors controlled by this worker.
func (w *Worker) Stop() error {
	w.cancelFunc()
	return nil
}

// EventSensorWorker is a worker that generates sensors on some kind of
// event(s).
type EventSensorWorker struct {
	EventSensorType
	Worker
}

func (w *EventSensorWorker) IsDisabled() bool {
	if w == nil {
		return true
	}

	if w.EventSensorType == nil {
		return true
	}

	return false
}

func (w *EventSensorWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handleSensorEvents(updatesCtx, w.EventSensorType), nil
}

func NewEventSensorWorker(id string) *EventSensorWorker {
	return &EventSensorWorker{
		Worker: Worker{WorkerID: id},
	}
}

// PollingSensorWorker is a worker that requires polling for generating sensors.
// It has a poll interval and jitter amount.
type PollingSensorWorker struct {
	PollingSensorType
	Worker
	PollInterval time.Duration
	JitterAmount time.Duration
}

func (w *PollingSensorWorker) IsDisabled() bool {
	if w == nil {
		return true
	}

	if w.PollingSensorType == nil {
		return true
	}

	return false
}

func (w *PollingSensorWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handleSensorPolling(updatesCtx, w.PollInterval, w.JitterAmount, w.PollingSensorType), nil
}

func NewPollingSensorWorker(id string, interval, jitter time.Duration) *PollingSensorWorker {
	return &PollingSensorWorker{
		Worker:       Worker{WorkerID: id},
		PollInterval: interval,
		JitterAmount: jitter,
	}
}

// OneShotSensorWorker is a worker that runs one-time, generates sensors, then
// exits.
type OneShotSensorWorker struct {
	OneShotSensorType
	Worker
}

func (w *OneShotSensorWorker) IsDisabled() bool {
	if w == nil {
		return true
	}

	if w.OneShotSensorType == nil {
		return true
	}

	return false
}

func (w *OneShotSensorWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handleSensorOneShot(updatesCtx, w.OneShotSensorType), nil
}

func NewOneShotSensorWorker(id string) *OneShotSensorWorker {
	return &OneShotSensorWorker{
		Worker: Worker{WorkerID: id},
	}
}

type EventWorker struct {
	EventType
	Worker
}

func (w *EventWorker) IsDisabled() bool {
	if w == nil {
		return true
	}

	if w.EventType == nil {
		return true
	}

	return false
}

func (w *EventWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handleEvents(updatesCtx, w.EventType), nil
}

func NewEventWorker(id string) *EventWorker {
	return &EventWorker{
		Worker: Worker{WorkerID: id},
	}
}

// handleSensorPolling: create an updater function to run the worker's Sensors
// function and pass this to the PollSensors helper, using the interval
// and jitter the worker has requested.
func handleSensorPolling(ctx context.Context, interval, jitter time.Duration, worker PollingSensorType) <-chan models.Entity {
	outCh := make(chan models.Entity)

	updater := func(d time.Duration) {
		// Send the delta (time since last poll) to the worker. Some workers may
		// not use this value and the UpdateDelta for them will be a no-op.
		worker.UpdateDelta(d)
		// Get the updated sensors.
		sensors, err := worker.Sensors(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker_type", "polling")).
				Error("Worker error occurred.", slog.Any("error", err))

			return
		}

		if len(sensors) == 0 {
			logging.FromContext(ctx).
				With(slog.String("worker_type", "polling")).
				Warn("Worker returned no sensors.")

			return
		}

		for _, s := range sensors {
			outCh <- s
		}
	}
	go func() {
		defer close(outCh)
		helpers.PollSensors(ctx, updater, interval, jitter)
	}()

	return outCh
}

// handleSensorEvents: read sensors from the worker Events function and pass these on.
func handleSensorEvents(ctx context.Context, worker EventSensorType) <-chan models.Entity {
	outCh := make(chan models.Entity)

	go func() {
		defer close(outCh)

		eventCh, err := worker.Events(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker_type", "events")).
				Debug("Unable to retrieve sensor events.", slog.Any("error", err))

			return
		}

		for s := range eventCh {
			outCh <- s
		}
	}()

	return outCh
}

// handleSensorOneShot: run the worker Sensors function to gather the sensors, pass these
// through the channel, then close it.
func handleSensorOneShot(ctx context.Context, worker OneShotSensorType) <-chan models.Entity {
	outCh := make(chan models.Entity)

	go func() {
		defer close(outCh)

		sensors, err := worker.Sensors(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker_type", "oneshot")).
				Debug("Unable to retrieve sensors.", slog.Any("error", err))

			return
		}

		for _, s := range sensors {
			outCh <- s
		}
	}()

	return outCh
}

// handleEvents: read events from the worker Events function and pass these on.
func handleEvents(ctx context.Context, worker EventType) <-chan models.Entity {
	outCh := make(chan models.Entity)

	go func() {
		defer close(outCh)

		eventCh, err := worker.Events(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker_type", "events")).
				Debug("Unable to retrieve sensor events.", slog.Any("error", err))

			return
		}

		for s := range eventCh {
			outCh <- s
		}
	}()

	return outCh
}
