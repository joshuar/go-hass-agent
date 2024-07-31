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

	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

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

// SensorController represents an object that manages one or more Workers.
type SensorController interface {
	// ActiveWorkers is a list of the names of all currently active Workers.
	ActiveWorkers() []string
	// InactiveWorkers is a list of the names of all currently inactive Workers.
	InactiveWorkers() []string
	// Start provides a way to start the named Worker.
	Start(ctx context.Context, name string) (<-chan sensor.Details, error)
	// Stop provides a way to stop the named Worker.
	Stop(name string) error
	// StartAll will start all Workers that this controller manages.
	StartAll(ctx context.Context) (<-chan sensor.Details, error)
	// StopAll will stop all Workers that this controller manages.
	StopAll() error
}

// Worker represents an object that is responsible for controlling the
// publishing of one or more sensors.
type Worker interface {
	ID() string
	// Sensors returns an array of the current value of all sensors, or a
	// non-nil error if this is not possible.
	Sensors(ctx context.Context) ([]sensor.Details, error)
	// Updates returns a channel on which updates to sensors will be published,
	// when they become available.
	Updates(ctx context.Context) (<-chan sensor.Details, error)
	// Stop is used to tell the worker to stop any background updates of
	// sensors.
	Stop() error
}

// MQTTController represents an object that is responsible for controlling the
// publishing of one or more commands over MQTT.
type MQTTController interface {
	// Subscriptions is a list of MQTT subscriptions this object wants to
	// establish on the MQTT broker.
	Subscriptions() []*mqttapi.Subscription
	// Configs are MQTT messages sent to the broker that Home Assistant will use
	// to set up entities.
	Configs() []*mqttapi.Msg
	// Msgs returns a channel on which this object will send MQTT messages on
	// certain events.
	Msgs() chan *mqttapi.Msg
}

type Controller interface {
	SensorController
	MQTTController
}

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

// runWorkers will call all the sensor worker functions that have been defined
// for this device.
func (agent *Agent) runWorkers(ctx context.Context, controllers ...SensorController) []<-chan sensor.Details {
	var sensorCh []<-chan sensor.Details

	for _, controller := range controllers {
		ch, err := controller.StartAll(ctx)
		if err != nil {
			agent.logger.Warn("Start controller had errors.", "errors", err.Error())
		} else {
			sensorCh = append(sensorCh, ch)
		}
	}

	if len(sensorCh) == 0 {
		agent.logger.Warn("No workers were started by any controllers.")

		return sensorCh
	}

	go func() {
		<-ctx.Done()

		for _, controller := range controllers {
			if err := controller.StopAll(); err != nil {
				agent.logger.Warn("Stop controller had errors.", "error", err.Error())
			}
		}
	}()

	return sensorCh
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

// runMQTTWorker will set up a connection to MQTT and listen on topics for
// controlling this device from Home Assistant.
func (agent *Agent) runMQTTWorker(ctx context.Context, controllers ...MQTTController) {
	var ( //nolint:prealloc
		subscriptions []*mqttapi.Subscription
		configs       []*mqttapi.Msg
		msgCh         []<-chan *mqttapi.Msg
		err           error
	)

	// Add the subscriptions and configs from the controllers.
	for _, controller := range controllers {
		subscriptions = append(subscriptions, controller.Subscriptions()...)
		configs = append(configs, controller.Configs()...)
		msgCh = append(msgCh, controller.Msgs())
	}

	// Create a new connection to the MQTT broker. This will also publish the
	// device subscriptions.
	client, err := mqttapi.NewClient(ctx, agent.prefs.GetMQTTPreferences(), subscriptions, configs)
	if err != nil {
		agent.logger.Error("Could not connect to MQTT.", "error", err.Error())

		return
	}

	go func() {
		agent.logger.Debug("Listening for messages to publish to MQTT.")

		for {
			select {
			case msg := <-mergeCh(ctx, msgCh...):
				if err := client.Publish(ctx, msg); err != nil {
					agent.logger.Warn("Unable to publish message to MQTT.", "topic", msg.Topic, "content", slog.Any("msg", msg.Message))
				}
			case <-ctx.Done():
				agent.logger.Debug("Stopped listening for messages to publish to MQTT.")

				return
			}
		}
	}()

	<-ctx.Done()
}

func (agent *Agent) resetMQTTWorker(ctx context.Context, osController MQTTController) error {
	if !agent.prefs.GetMQTTPreferences().IsMQTTEnabled() {
		return nil
	}

	client, err := mqttapi.NewClient(ctx, agent.prefs.GetMQTTPreferences(), nil, nil)
	if err != nil {
		return fmt.Errorf("could not connect to MQTT: %w", err)
	}

	if err := client.Unpublish(ctx, osController.Configs()...); err != nil {
		return fmt.Errorf("could not remove configs from MQTT: %w", err)
	}

	return nil
}
