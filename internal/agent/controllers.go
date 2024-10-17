// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"fmt"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type WorkerController[T any] interface {
	ID() string
	ActiveWorkers() []string
	// InactiveWorkers is a list of the names of all currently inactive Workers.
	InactiveWorkers() []string
	// Start provides a way to start the named Worker.
	Start(ctx context.Context, name string) (<-chan T, error)
	// Stop provides a way to stop the named Worker.
	Stop(name string) error
}

// SensorController represents an object that manages one or more Workers.
type SensorController interface {
	WorkerController[sensor.Entity]
	// States returns the list of all sensor states tracked by all workers of
	// this controller.
	States(ctx context.Context) []sensor.Entity
}

type EventController interface {
	WorkerController[event.Event]
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

func (agent *Agent) setupControllers(ctx context.Context, prefs *preferences.Preferences) []any {
	var (
		mqttDevice  *mqtthass.Device
		controllers []any
	)

	// If MQTT functionality is enabled create an MQTT device, used to configure
	// MQTT functionality for some controllers.
	if prefs.IsMQTTEnabled() {
		mqttDevice = prefs.GenerateMQTTDevice(ctx)
		// Create an MQTT commands controller.
		mqttCmdController := newMQTTController(ctx, mqttDevice)
		if mqttCmdController != nil {
			controllers = append(controllers, mqttCmdController)
		}
		// Add the OS MQTT controller.
		controllers = append(controllers, newOSMQTTController(ctx, mqttDevice))
	}

	scriptsController := newScriptsController(ctx)
	if scriptsController != nil {
		controllers = append(controllers, scriptsController)
	}

	// Create a new device controller. The controller will have all the
	// necessary configuration for device-specific sensors.
	devController := agent.newDeviceController(ctx, prefs)
	if devController != nil {
		controllers = append(controllers, devController)
	}
	// Create a new OS controller. The controller will have all the necessary
	// configuration for any OS-specific sensors.
	osSensorControllers, osEventControllers := newOperatingSystemControllers(ctx)
	controllers = append(controllers, osSensorControllers, osEventControllers)

	return controllers
}

func runControllerWorkers[T any](ctx context.Context, prefs *preferences.Preferences, controllers ...WorkerController[T]) {
	// Start all inactive workers of all controllers.
	eventCh := startAllWorkers(ctx, controllers)
	if len(eventCh) == 0 {
		logging.FromContext(ctx).Warn("No workers were started by any controllers.")
		return
	}

	hassclient, err := newHassClient(ctx, prefs)
	if err != nil {
		logging.FromContext(ctx).Error("Cannot start workers.", slog.Any("error", err))
		return
	}

	// When the context is done, stop all active workers of all controllers.
	go func() {
		<-ctx.Done()
		stopAllWorkers(ctx, controllers)
	}()

	// Process all events/sensors from all workers.
	for details := range mergeCh(ctx, eventCh...) {
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

func startAllWorkers[T any](ctx context.Context, controllers []WorkerController[T]) []<-chan T {
	var eventCh []<-chan T

	for _, controller := range controllers {
		logging.FromContext(ctx).Debug("Starting controller",
			slog.String("controller", controller.ID()))

		for _, workerName := range controller.InactiveWorkers() {
			logging.FromContext(ctx).Debug("Starting worker",
				slog.String("worker", workerName))

			workerCh, err := controller.Start(ctx, workerName)
			if err != nil {
				logging.FromContext(ctx).
					Warn("Could not start worker.",
						slog.String("controller", controller.ID()),
						slog.String("worker", workerName),
						slog.Any("errors", err))
			} else {
				eventCh = append(eventCh, workerCh)
			}
		}
	}

	return eventCh
}

func stopAllWorkers[T any](ctx context.Context, controllers []WorkerController[T]) {
	for _, controller := range controllers {
		logging.FromContext(ctx).Debug("Stopping controller", slog.String("controller", controller.ID()))

		for _, workerName := range controller.ActiveWorkers() {
			logging.FromContext(ctx).Debug("Stopping worker", slog.String("worker", workerName))

			if err := controller.Stop(workerName); err != nil {
				logging.FromContext(ctx).
					Warn("Could not stop worker.",
						slog.String("controller", controller.ID()),
						slog.String("worker", workerName),
						slog.Any("errors", err))
			}
		}
	}
}

func newHassClient(ctx context.Context, prefs *preferences.Preferences) (*hass.Client, error) {
	client, err := hass.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create Home Assistant client: %w", err)
	}

	client.Endpoint(prefs.RestAPIURL(), hass.DefaultTimeout)

	return client, nil
}
