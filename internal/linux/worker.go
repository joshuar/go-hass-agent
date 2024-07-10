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
	Value    any
	WorkerID string
}

// Name is the name of the group of sensors managed by this SensorWorker.
func (w *SensorWorker) ID() string {
	return w.WorkerID
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
			return nil, fmt.Errorf("failed to get current state of polling sensors: %w", err)
		}

		return sensors, nil
	case oneShotType:
		sensors, err := worker.Sensors(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current state of polling sensors: %w", err)
		}

		return sensors, nil
	}

	return nil, ErrUnknownWorker
}

// Updates returns a channel on which sensor updates can be received. If the
// functionality to send sensor updates cannot be achieved, it will return a
// non-nil error.
//
//nolint:cyclop
//revive:disable:function-length
func (w *SensorWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	outCh := make(chan sensor.Details)

	var workerLogAttrs []any

	workerLogAttrs = append(workerLogAttrs, slog.String("id", w.WorkerID))

	switch worker := w.Value.(type) {
	case pollingType:
		workerLogAttrs = append(workerLogAttrs, slog.String("worker_type", "polling"))

		// pollingType: create an updater function to run the worker's Sensors
		// function and pass this to the PollSensors helper, using the interval
		// and jitter the worker has requested.
		logging.FromContext(ctx).Debug("Starting worker.", workerLogAttrs...)

		updater := func(d time.Duration) {
			sensors, err := worker.Sensors(ctx, d)
			if err != nil {
				logging.FromContext(ctx).Warn("Unable to retrieve sensors.", workerLogAttrs...)

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
	case eventType:
		workerLogAttrs = append(workerLogAttrs, slog.String("worker_type", "event"))

		// eventType: read sensors from the worker Events function and pass
		// these on.
		go func() {
			defer close(outCh)

			eventCh, err := worker.Events(ctx)
			if err != nil {
				workerLogAttrs = append(workerLogAttrs, slog.Any("error", err))
				logging.FromContext(ctx).Debug("Could not start worker.", workerLogAttrs...)

				return
			}

			logging.FromContext(ctx).Debug("Starting worker.", workerLogAttrs...)

			for s := range eventCh {
				outCh <- s
			}
		}()
	case oneShotType:
		workerLogAttrs = append(workerLogAttrs, slog.String("worker_type", "one-shot"))

		// oneShot: run the worker Sensors function to gather the sensors, pass
		// these through the channel, then close it.
		go func() {
			defer close(outCh)

			sensors, err := worker.Sensors(ctx)
			if err != nil {
				workerLogAttrs = append(workerLogAttrs, slog.Any("error", err))
				logging.FromContext(ctx).Debug("Unable to retrieve sensors.", workerLogAttrs...)

				return
			}

			logging.FromContext(ctx).Debug("Starting worker.", workerLogAttrs...)

			for _, s := range sensors {
				outCh <- s
			}
		}()
	default:
		// default: we should not get here, so if we do, return an error
		// indicating we don't know what type of worker this is.
		return nil, fmt.Errorf("could not track updates for %s: %w", w.WorkerID, ErrUnknownWorker)
	}

	return outCh, nil
}
