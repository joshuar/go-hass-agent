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
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

var ErrUnknownWorker = errors.New("unknown sensor worker type")

// pollingType interface represents sensors that are generated on some poll interval.
type pollingType interface {
	Interval() time.Duration
	Jitter() time.Duration
	Sensors(ctx context.Context, delta time.Duration) ([]sensor.Details, error)
}

// dbusType interface represents sensors that are generated on D-Bus events.
type dbusType interface {
	Setup(ctx context.Context) *dbusx.Watch
	Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details
	Sensors(ctx context.Context) ([]sensor.Details, error)
}

// eventType interface represents sensors that are generated by some means other
// than D-Bus signals.
type eventType interface {
	Events(ctx context.Context) chan sensor.Details
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
	case dbusType:
		sensors, err := worker.Sensors(ctx)
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
func (w *SensorWorker) Updates(ctx context.Context) (<-chan sensor.Details, error) {
	outCh := make(chan sensor.Details)

	switch worker := w.Value.(type) {
	case pollingType:
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
			helpers.PollSensors(ctx, updater, worker.Interval(), worker.Jitter())
		}()
	case dbusType:
		eventCh, err := dbusx.WatchBus(ctx, worker.Setup(ctx))
		if err != nil {
			close(outCh)
			return outCh, fmt.Errorf("could not set up watch for worker: %w", err)
		}
		go func() {
			defer close(outCh)

			for s := range worker.Watch(ctx, eventCh) {
				outCh <- s
			}
		}()
	case eventType:
		go func() {
			defer close(outCh)

			for s := range worker.Events(ctx) {
				outCh <- s
			}
		}()
	case oneShotType:
		go func() {
			defer close(outCh)
			sensors, err := worker.Sensors(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("Unable to retrieve sensors.")
				return
			}
			for _, s := range sensors {
				outCh <- s
			}
		}()
	default:
		return nil, ErrUnknownWorker
	}
	return outCh, nil
}
