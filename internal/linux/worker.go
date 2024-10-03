// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run github.com/matryer/moq -out worker_mocks_test.go . PollingType EventType OneShotType
package linux

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

var ErrUnknownWorker = errors.New("unknown sensor worker type")

// PollingType interface represents sensors that are generated on some poll interval.
type PollingType interface {
	UpdateDelta(delta time.Duration)
	Sensors(ctx context.Context) ([]sensor.Entity, error)
}

// EventType interface represents sensors that are generated on some event
// trigger, such as D-Bus messages.
type EventType interface {
	Events(ctx context.Context) (<-chan sensor.Entity, error)
	Sensors(ctx context.Context) ([]sensor.Entity, error)
}

// OneShotType interface represents sensors that are generated only one-time and
// have no ongoing updates.
type OneShotType interface {
	Sensors(ctx context.Context) ([]sensor.Entity, error)
}

// SensorWorker is a struct embedded in other specific worker structs. It holds
// a function to cancel the worker and its ID.
type SensorWorker struct {
	cancelFunc context.CancelFunc
	WorkerID   string
}

// ID is a name that can be used as an ID to represent the group of sensors
// managed by this worker.
func (w *SensorWorker) ID() string {
	if w != nil {
		return w.WorkerID
	}

	return "Unknown Worker"
}

// Stop will stop any processing of sensors controlled by this worker.
func (w *SensorWorker) Stop() error {
	slog.Debug("Stopping worker", slog.String("worker", w.ID()))
	w.cancelFunc()

	return nil
}

// EventSensorWorker is a worker that generates sensors on some kind of
// event(s).
type EventSensorWorker struct {
	EventType
	SensorWorker
}

func (w *EventSensorWorker) Start(ctx context.Context) (<-chan sensor.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handleEvents(updatesCtx, w.EventType), nil
}

func NewEventWorker(id string) *EventSensorWorker {
	return &EventSensorWorker{
		SensorWorker: SensorWorker{WorkerID: id},
	}
}

// PollingSensorWorker is a worker that requires polling for generating sensors.
// It has a poll interval and jitter amount.
type PollingSensorWorker struct {
	PollingType
	SensorWorker
	PollInterval time.Duration
	JitterAmount time.Duration
}

func (w *PollingSensorWorker) Start(ctx context.Context) (<-chan sensor.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handlePolling(updatesCtx, w.PollInterval, w.JitterAmount, w.PollingType), nil
}

func NewPollingWorker(id string, interval, jitter time.Duration) *PollingSensorWorker {
	return &PollingSensorWorker{
		SensorWorker: SensorWorker{WorkerID: id},
		PollInterval: interval,
		JitterAmount: jitter,
	}
}

// OneShotSensorWorker is a worker that runs one-time, generates sensors, then
// exits.
type OneShotSensorWorker struct {
	OneShotType
	SensorWorker
}

func (w *OneShotSensorWorker) Start(ctx context.Context) (<-chan sensor.Entity, error) {
	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	return handleOneShot(updatesCtx, w.OneShotType), nil
}

func NewOneShotWorker(id string) *OneShotSensorWorker {
	return &OneShotSensorWorker{
		SensorWorker: SensorWorker{WorkerID: id},
	}
}

// handlePolling: create an updater function to run the worker's Sensors
// function and pass this to the PollSensors helper, using the interval
// and jitter the worker has requested.
func handlePolling(ctx context.Context, interval, jitter time.Duration, worker PollingType) <-chan sensor.Entity {
	outCh := make(chan sensor.Entity)

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

// handleEvents: read sensors from the worker Events function and pass these on.
func handleEvents(ctx context.Context, worker EventType) <-chan sensor.Entity {
	outCh := make(chan sensor.Entity)

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

// handleOneShot: run the worker Sensors function to gather the sensors, pass these
// through the channel, then close it.
func handleOneShot(ctx context.Context, worker OneShotType) <-chan sensor.Entity {
	outCh := make(chan sensor.Entity)

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
