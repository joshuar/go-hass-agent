// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//go:generate moq -out runners_mocks_test.go . SensorController Worker Script LocationUpdateResponse SensorUpdateResponse SensorRegistrationResponse
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/robfig/cron/v3"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

var scriptPath string

var (
	ErrStateUpdateUnknown = errors.New("unknown sensor update response")
	ErrStateUpdateFailed  = errors.New("state update failed")
	ErrRegDisableFailed   = errors.New("failed to disable sensor in registry")
	ErrRegAddFailed       = errors.New("failed to set registered status for sensor in registry")
	ErrTrkUpdateFailed    = errors.New("failed to update sensor state in tracker")
	ErrRegistrationFailed = errors.New("sensor registration failed")
)

type Script interface {
	Schedule() string
	Execute() (sensors []sensor.Details, err error)
}

type LocationUpdateResponse interface {
	Updated() bool
	Error() string
}

type SensorUpdateResponse interface {
	Updates() map[string]*sensor.UpdateStatus
}

type SensorRegistrationResponse interface {
	Registered() bool
	Error() string
}

// runScripts will retrieve all scripts that the agent can run and queue them up
// to be run on their defined schedule using the cron scheduler. It also sets up
// a channel to receive script output and send appropriate sensor objects to the
// sensor.
//
//revive:disable:function-length
func (agent *Agent) runScripts(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	// Define the path to custom sensor scripts.
	if scriptPath == "" {
		scriptPath = filepath.Join(xdg.ConfigHome, agent.id, "scripts")
	}
	// Get any scripts in the script path.
	sensorScripts, err := findScripts(scriptPath)

	// If no scripts were found or there was an error processing scripts, log a
	// message and return.
	switch {
	case err != nil:
		agent.logger.Warn("Error finding custom sensor scripts.", "error", err.Error())
		close(sensorCh)

		return sensorCh
	case len(sensorScripts) == 0:
		agent.logger.Debug("No custom sensor scripts found.")
		close(sensorCh)

		return sensorCh
	}

	scheduler := cron.New()

	var jobs []cron.EntryID //nolint:prealloc // we can't determine size in advance.

	for _, script := range sensorScripts {
		if script.Schedule() == "" {
			agent.logger.Warn("Script found without schedule defined, skipping.")

			continue
		}
		// Create a closure to run the script on it's schedule.
		runFunc := func() {
			sensors, err := script.Execute()
			if err != nil {
				agent.logger.Warn("Could not execute script.", "error", err.Error())

				return
			}

			for _, o := range sensors {
				sensorCh <- o
			}
		}
		// Add the script to the cron scheduler to run the closure on it's
		// defined schedule.
		jobID, err := scheduler.AddFunc(script.Schedule(), runFunc)
		if err != nil {
			agent.logger.Warn("Unable to schedule script", "error", err.Error())

			break
		}

		jobs = append(jobs, jobID)
	}

	agent.logger.Debug("Starting cron scheduler for script sensors.")
	// Start the cron scheduler
	scheduler.Start()

	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		// Stop next run of all active cron jobs.
		for _, jobID := range jobs {
			scheduler.Remove(jobID)
		}

		agent.logger.Debug("Stopping cron scheduler for script sensors.")
		// Stop the scheduler once all jobs have finished.
		cronCtx := scheduler.Stop()
		<-cronCtx.Done()
	}()

	return sensorCh
}

func (agent *Agent) processSensors(ctx context.Context, trk SensorTracker, reg Registry, sensorCh ...<-chan sensor.Details) {
	client := hass.NewDefaultHTTPClient(agent.prefs.Hass.RestAPIURL)

	for update := range mergeCh(ctx, sensorCh...) {
		go func(upd sensor.Details) {
			// Ignore disabled sensors.
			if reg.IsDisabled(upd.ID()) {
				return
			}

			var (
				req  hass.PostRequest
				resp hass.Response
				err  error
			)

			// Create the request/response objects depending on what kind of
			// sensor update this is.
			if _, ok := upd.State().(*sensor.LocationRequest); ok {
				req, resp, err = sensor.NewLocationUpdateRequest(upd)
			} else {
				if reg.IsRegistered(upd.ID()) {
					req, resp, err = sensor.NewUpdateRequest(upd)
				} else {
					req, resp, err = sensor.NewRegistrationRequest(upd)
				}
			}

			// If there was an error creating the request/response objects,
			// abort.
			if err != nil {
				agent.logger.Warn("Could not create sensor update request.", "sensor_id", upd.ID(), "error", err.Error())

				return
			}

			// Send the request/response details to Home Assistant.
			err = hass.ExecuteRequest(ctx, client, "", req, resp)
			if err != nil {
				agent.logger.Warn("Failed to send sensor details to Home Assistant.", "sensor_id", upd.ID(), "error", err.Error())

				return
			}

			// Handle the response received.
			agent.processResponse(ctx, upd, resp, reg, trk)
		}(update)
	}
}

func (agent *Agent) processResponse(ctx context.Context, upd sensor.Details, resp any, reg Registry, trk SensorTracker) {
	sensorLogAttrs := slog.Group("sensor",
		slog.String("name", upd.Name()),
		slog.String("id", upd.ID()),
		slog.Any("state", upd.State()),
		slog.String("units", upd.Units()))

	switch details := resp.(type) {
	case LocationUpdateResponse:
		if details.Updated() {
			agent.logger.LogAttrs(ctx, slog.LevelDebug, "Location updated.")
		} else {
			agent.logger.Warn("Location update failed.", slog.String("error", details.Error()))
		}
	case SensorUpdateResponse:
		for _, status := range details.Updates() {
			success, errs := processStateUpdates(trk, reg, upd, status)
			if !success {
				agent.logger.Warn("Sensor update failed.", sensorLogAttrs, slog.Any("error", errs))
			} else {
				if errs != nil {
					agent.logger.LogAttrs(ctx, slog.LevelDebug, "Sensor update succeeded with warnings.", sensorLogAttrs, slog.Any("warnings", errs))
				} else {
					agent.logger.LogAttrs(ctx, slog.LevelDebug, "Sensor updated.", sensorLogAttrs)
				}
			}
		}
	case SensorRegistrationResponse:
		success, errs := processRegistration(trk, reg, upd, details)
		if !success {
			agent.logger.Warn("Sensor registration failed.", sensorLogAttrs, slog.Any("error", errs))
		} else {
			if errs != nil {
				agent.logger.LogAttrs(ctx, slog.LevelDebug, "Sensor registration succeeded with warnings.", sensorLogAttrs, slog.Any("warnings", errs))
			} else {
				agent.logger.LogAttrs(ctx, slog.LevelDebug, "Sensor registered.", sensorLogAttrs)
			}
		}
	}
}

func processStateUpdates(trk SensorTracker, reg Registry, upd sensor.Details, status *sensor.UpdateStatus) (bool, error) {
	// No status was returned.
	if status == nil {
		return false, ErrStateUpdateUnknown
	}
	// The update failed.
	if !status.Success {
		if status.Error != nil {
			return false, fmt.Errorf("%w, code %d: reason: %s", ErrStateUpdateFailed, status.Error.Code, status.Error.Message)
		}

		return false, fmt.Errorf("%w, unknown reason", ErrStateUpdateFailed)
	}

	// At this point, the sensor update was successful. Any errors are really
	// warnings and non-critical.
	var warnings error

	// If HA reports the sensor as disabled, update the registry.
	if reg.IsDisabled(upd.ID()) != status.Disabled {
		if err := reg.SetDisabled(upd.ID(), status.Disabled); err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrRegDisableFailed, err))
		}
	}

	// Add the sensor update to the tracker.
	if err := trk.Add(upd); err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrTrkUpdateFailed, err))
	}

	// Return success status and any warnings.
	return true, warnings
}

func processRegistration(trk SensorTracker, reg Registry, upd sensor.Details, details SensorRegistrationResponse) (bool, error) {
	// If the registration failed, log a warning.
	if !details.Registered() {
		return false, fmt.Errorf("%w: %s", ErrRegistrationFailed, details.Error())
	}

	// At this point, the sensor registration was successful. Any errors are really
	// warnings and non-critical.
	var warnings error

	// Set the sensor as registered in the registry.
	err := reg.SetRegistered(upd.ID(), true)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrRegAddFailed, err))
	}
	// Update the sensor state in the tracker.
	if err := trk.Add(upd); err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("%w: %w", ErrTrkUpdateFailed, err))
	}

	// Return success status and any warnings.
	return true, warnings
}
