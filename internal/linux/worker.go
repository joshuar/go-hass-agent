// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

var ErrUnknownWorker = errors.New("unknown sensor worker type")

// pollingType interface represents sensors that are generated on some poll interval.
type pollingType interface {
	Interval() time.Duration
	Jitter() time.Duration
	Sensors(ctx context.Context, delta time.Duration) ([]sensor.Details, error)
}

// eventType interface represents sensors that are generated on some event
// trigger, such as D-Bus messages.
type eventType interface {
	Events(ctx context.Context) (chan sensor.Details, error)
	Sensors(ctx context.Context) ([]sensor.Details, error)
}

// oneShotType interface represents sensors that are generated only one-time and
// have no ongoing updates.
type oneShotType interface {
	Sensors(ctx context.Context) ([]sensor.Details, error)
}

// SensorWorker represents the functionality to track a group of one or more
// related sensors.
type SensorWorker struct {
	Value      any
	cancelFunc context.CancelFunc
	logger     *slog.Logger
	WorkerID   string
}

// ID is a name that can be used as an ID to represent the group of sensors
// managed by this worker.
func (w *SensorWorker) ID() string {
	return w.WorkerID
}

// Stop will stop any processing of sensors controlled by this worker.
func (w *SensorWorker) Stop() error {
	w.cancelFunc()

	return nil
}

// Sensors returns the current values of all sensors managed by this
// SensorWorker. If the values cannot be retrieved, it will return a non-nil
// error.
func (w *SensorWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	switch worker := w.Value.(type) {
	case pollingType:
		sensors, err := worker.Sensors(ctx, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get current state of polling sensors: %w", err)
		}

		return sensors, nil
	case eventType:
		sensors, err := worker.Sensors(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current state of event sensors: %w", err)
		}

		return sensors, nil
	case oneShotType:
		sensors, err := worker.Sensors(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current state of one-shot sensors: %w", err)
		}

		return sensors, nil
	}

	return nil, ErrUnknownWorker
}

// Updates returns a channel on which sensor updates can be received. If the
// functionality to send sensor updates cannot be achieved, it will return a
// non-nil error.
func (w *SensorWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	var outCh chan sensor.Details

	// Create a new context for the updates scope.
	updatesCtx, cancelFunc := context.WithCancel(ctx)
	// Save the context cancelFunc in the worker to be used as part of its
	// Stop() method.
	w.cancelFunc = cancelFunc
	// Create a child logger for the worker.

	// Handle the worker appropriately based on its type.
	switch worker := w.Value.(type) {
	case pollingType:
		w.logger = logging.FromContext(ctx).With(
			slog.String("worker", w.ID()),
			slog.String("worker_type", "polling"))
		outCh = w.handlePolling(updatesCtx, worker)
	case eventType:
		w.logger = logging.FromContext(ctx).With(
			slog.String("worker", w.ID()),
			slog.String("worker_type", "polling"))
		outCh = w.handleEvents(updatesCtx, worker)
	case oneShotType:
		w.logger = logging.FromContext(ctx).With(
			slog.String("worker", w.ID()),
			slog.String("worker_type", "polling"))
		outCh = w.handleOneShot(updatesCtx, worker)
	default:
		// default: we should not get here, so if we do, return an error
		// indicating we don't know what type of worker this is.
		return nil, fmt.Errorf("could not track updates for %s: %w", w.WorkerID, ErrUnknownWorker)
	}

	return outCh, nil
}

// handlePolling: create an updater function to run the worker's Sensors
// function and pass this to the PollSensors helper, using the interval
// and jitter the worker has requested.
func (w *SensorWorker) handlePolling(ctx context.Context, worker pollingType) chan sensor.Details {
	outCh := make(chan sensor.Details)

	updater := func(d time.Duration) {
		sensors, err := worker.Sensors(ctx, d)
		if err != nil {
			w.logger.Error("Worker error occurred.", slog.Any("error", err))
			return
		}

		if len(sensors) == 0 {
			w.logger.Warn("Worker returned no sensors.")
			return
		}

		for _, s := range sensors {
			outCh <- s
		}
	}
	go func() {
		defer close(outCh)
		helpers.PollSensors(ctx, updater, worker.Interval(), worker.Jitter())
	}()

	return outCh
}

// handleEvents: read sensors from the worker Events function and pass these on.
func (w *SensorWorker) handleEvents(ctx context.Context, worker eventType) chan sensor.Details {
	outCh := make(chan sensor.Details)

	go func() {
		defer close(outCh)

		eventCh, err := worker.Events(ctx)
		if err != nil {
			w.logger.Debug("Unable to retrieve sensor events.", slog.Any("error", err))

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
func (w *SensorWorker) handleOneShot(ctx context.Context, worker oneShotType) chan sensor.Details {
	outCh := make(chan sensor.Details)

	go func() {
		defer close(outCh)

		sensors, err := worker.Sensors(ctx)
		if err != nil {
			w.logger.Debug("Unable to retrieve sensors.", slog.Any("error", err))

			return
		}

		for _, s := range sensors {
			outCh <- s
		}
	}()

	return outCh
}
