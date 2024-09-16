// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type worker interface {
	ID() string
	Start(ctx context.Context) (<-chan sensor.Details, error)
	Sensors(ctx context.Context) ([]sensor.Details, error)
	Stop() error
}

type workerState struct {
	worker
	started bool
}

type sensorController struct {
	workers map[string]*workerState
	id      string
}

func (c *sensorController) ID() string {
	return c.id
}

func (c *sensorController) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, len(c.workers))

	for id, worker := range c.workers {
		if worker.started {
			activeWorkers = append(activeWorkers, id)
		}
	}

	return activeWorkers
}

func (c *sensorController) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, len(c.workers))

	for id, worker := range c.workers {
		if !worker.started {
			inactiveWorkers = append(inactiveWorkers, id)
		}
	}

	return inactiveWorkers
}

func (c *sensorController) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	worker, exists := c.workers[name]
	if !exists {
		return nil, ErrUnknownWorker
	}

	if worker.started {
		return nil, ErrWorkerAlreadyStarted
	}

	workerCh, err := c.workers[name].Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}

	c.workers[name].started = true

	return workerCh, nil
}

func (c *sensorController) Stop(name string) error {
	// Check if the given worker ID exists.
	worker, exists := c.workers[name]
	if !exists {
		return ErrUnknownWorker
	}
	// Stop the worker. Report any errors.
	if err := worker.Stop(); err != nil {
		return fmt.Errorf("error stopping worker: %w", err)
	}

	return nil
}

func (c *sensorController) States(ctx context.Context) []sensor.Details {
	var sensors []sensor.Details

	for _, worker := range c.ActiveWorkers() {
		workerSensors, err := c.workers[worker].Sensors(ctx)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("controller", c.ID())).
				Debug("Could not retrieve worker sensors",
					slog.String("worker", c.workers[worker].ID()),
					slog.Any("error", err))
		}

		sensors = append(sensors, workerSensors...)
	}

	return sensors
}

// runSensorWorkers will start all the sensor worker functions for all sensor
// controllers passed in. It returns a single merged channel of sensor updates.
//
//nolint:gocognit
func runSensorWorkers(ctx context.Context, prefs *preferences.Preferences, controllers ...SensorController) {
	var sensorCh []<-chan sensor.Details

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
				sensorCh = append(sensorCh, workerCh)
			}
		}
	}

	if len(sensorCh) == 0 {
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
		logging.FromContext(ctx).Debug("Stopping all sensor controllers.")

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

	for details := range mergeCh(ctx, sensorCh...) {
		go func(details sensor.Details) {
			if err := hassclient.ProcessSensor(ctx, details); err != nil {
				logging.FromContext(ctx).Error("Process sensor failed.", slog.Any("error", err))
			}
		}(details)
	}
}
