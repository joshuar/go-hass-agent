// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
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
	// Value is a pointer to an interface that exposes methods to retrieve the
	// sensor values for this worker.
	Value any
	// WorkerName is a short name to refer to this group of sensors.
	WorkerName string
	// WorkerDesc describes what the sensors measure.
	WorkerDesc string
}

// Name is the name of the group of sensors managed by this SensorWorker.
func (w *SensorWorker) Name() string {
	return w.WorkerName
}

// Description is an explanation of what the sensors measure.
func (w *SensorWorker) Description() string {
	return w.WorkerDesc
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

	switch worker := w.Value.(type) {
	case pollingType:
		// pollingType: create an updater function to run the worker's Sensors
		// function and pass this to the PollSensors helper, using the interval
		// and jitter the worker has requested.
		updater := func(d time.Duration) {
			sensors, err := worker.Sensors(ctx, d)
			if err != nil {
				log.Warn().Err(err).Msg("Unable to retrieve sensors.")

				return
			}

			for _, s := range sensors {
				outCh <- s
			}
		}
		go func() {
			defer close(outCh)
			log.Trace().Str("worker", w.Name()).Msg("Polling for sensor updates...")
			helpers.PollSensors(ctx, updater, worker.Interval(), worker.Jitter())
		}()
	case eventType:
		// eventType: read sensors from the worker Events function and pass
		// these on.
		go func() {
			defer close(outCh)

			eventCh, err := worker.Events(ctx)
			if err != nil {
				log.Debug().Err(err).Msg("Could not start event worker.")

				return
			}

			log.Trace().Str("worker", w.Name()).Msg("Listening for sensor update events...")

			for s := range eventCh {
				outCh <- s
			}
		}()
	case oneShotType:
		// oneShot: run the worker Sensors function to gather the sensors, pass
		// these through the channel, then close it.
		go func() {
			defer close(outCh)

			sensors, err := worker.Sensors(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("Unable to retrieve sensors.")

				return
			}

			log.Trace().Str("worker", w.Name()).Msg("Sending sensors...")

			for _, s := range sensors {
				outCh <- s
			}
		}()
	default:
		// default: we should not get here, so if we do, return an error
		// indicating we don't know what type of worker this is.
		return nil, fmt.Errorf("could not track updates for %s: %w", w.Name(), ErrUnknownWorker)
	}

	return outCh, nil
}
