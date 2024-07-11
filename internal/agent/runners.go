// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"

	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/scripts"
)

// WorkerController represents an object that manages one or more Workers.
type WorkerController interface {
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

// MQTTWorker represents an object that is responsible for controlling the
// publishing of one or more commands over MQTT.
type MQTTWorker interface {
	// Subscriptions is a list of MQTT subscriptions this object wants to
	// establish on the MQTT broker.
	Subscriptions() []*mqttapi.Subscription
	// Configs are MQTT messages sent to the broker that Home Assistant will use
	// to set up entities.
	Configs() []*mqttapi.Msg
	// Msgs returns a channel on which this object will send MQTT messages on
	// certain events.
	Msgs() chan *mqttapi.Msg
	// Setup provides a way to perform any required initialisation of the
	// object.
	Setup(_ context.Context) error
}

// runWorkers will call all the sensor worker functions that have been defined
// for this device.
func runWorkers(ctx context.Context, trk SensorTracker, reg sensor.Registry) {
	// Create sensor workers for OS.
	osWorkers := createSensorWorkers(ctx)
	// Start sensor workers for OS.
	sensorUpdates, err := osWorkers.StartAll(ctx)
	if err != nil {
		logging.FromContext(ctx).Warn("Some sensor workers could not be started", "errors", err.Error())
	}
	// Create sensor workers for device.
	deviceWorkers := device.CreateSensorWorkers(ctx)
	// Start sensor workers for device.
	deviceUpdates, err := deviceWorkers.StartAll(ctx)
	if err != nil {
		logging.FromContext(ctx).Warn("Some device workers could not be started", "errors", err.Error())
	}

	// Listen for sensor updates from all workers.
	go func() {
		if err := trk.Process(ctx, reg, sensorUpdates, deviceUpdates); err != nil {
			logging.FromContext(ctx).Error("Could not process sensor updates", "error", err.Error())
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)

	// When context is cancelled, stop all sensor workers.
	go func() {
		defer wg.Done()
		<-ctx.Done()

		if err := osWorkers.StopAll(); err != nil {
			logging.FromContext(ctx).Debug("Error occurred trying to stop sensor workers.", "error", err.Error())
		}

		if err := deviceWorkers.StopAll(); err != nil {
			logging.FromContext(ctx).Debug("Error occurred trying to stop device workers.", "error", err.Error())
		}
	}()
	wg.Wait()
}

// runScripts will retrieve all scripts that the agent can run and queue them up
// to be run on their defined schedule using the cron scheduler. It also sets up
// a channel to receive script output and send appropriate sensor objects to the
// sensor.
func runScripts(ctx context.Context, path string, trk SensorTracker, reg sensor.Registry) {
	allScripts, err := scripts.FindScripts(ctx, path)

	switch {
	case err != nil:
		logging.FromContext(ctx).Warn("Error finding custom sensor scripts.", "error", err.Error())

		return
	case len(allScripts) == 0:
		logging.FromContext(ctx).Debug("No custom sensor scripts found.")

		return
	}

	scheduler := cron.New()

	outCh := make([]<-chan sensor.Details, 0, len(allScripts))

	for _, script := range allScripts {
		schedule := script.Schedule()
		if schedule != "" {
			_, err := scheduler.AddJob(schedule, script)
			if err != nil {
				logging.FromContext(ctx).Warn("Unable to schedule script", "script", script.Path(), "error", err.Error())

				break
			}

			outCh = append(outCh, script.Output)
			logging.FromContext(ctx).Debug("Script sensor scheduled.", "script", script.Path())
		}
	}

	logging.FromContext(ctx).Debug("Starting cron scheduler for script sensors.")
	scheduler.Start()

	go func() {
		if err := trk.Process(ctx, reg, outCh...); err != nil {
			logging.FromContext(ctx).Error("Could not process script sensor updates", "error", err.Error())
		}
	}()

	<-ctx.Done()
	logging.FromContext(ctx).Debug("Stopping cron scheduler for script sensors.")

	cronCtx := scheduler.Stop()
	<-cronCtx.Done()
}

// runNotificationsWorker will run a goroutine that is listening for
// notification messages from Home Assistant on a websocket connection. Any
// received notifications will be dipslayed on the device running the agent.
func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	notifyCh, err := hass.StartWebsocket(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Could not listen for notifications.", "error", err.Error())
	}

	logging.FromContext(ctx).Debug("Listening for notifications.")

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				logging.FromContext(ctx).Debug("Stopping notification handler.")

				return
			case n := <-notifyCh:
				agent.ui.DisplayNotification(n)
			}
		}
	}()

	wg.Wait()
}

// runMQTTWorker will set up a connection to MQTT and listen on topics for
// controlling this device from Home Assistant.
func (agent *Agent) runMQTTWorker(ctx context.Context, commandsFile string) {
	if !agent.prefs.MQTTEnabled {
		return
	}

	mqttCtx, mqttCancel := context.WithCancel(ctx)
	defer mqttCancel()

	var deviceController, commandController MQTTWorker

	var subscriptions []*mqttapi.Subscription

	var configs []*mqttapi.Msg

	var err error

	// Create an MQTT device for this operating system and run its Setup.
	deviceController = newMQTTDevice(mqttCtx)
	if err = deviceController.Setup(mqttCtx); err != nil {
		logging.FromContext(ctx).Error("Could not set up MQTT.", "error", err.Error())
	} else {
		subscriptions = append(subscriptions, deviceController.Subscriptions()...)
		configs = append(configs, deviceController.Configs()...)
	}

	// Create an MQTT device for this operating system and run its Setup.
	commandController, err = commands.NewCommandsController(ctx, commandsFile, device.MQTTDeviceInfo(ctx))
	if err != nil {
		logging.FromContext(ctx).Warn("Could not set up MQTT commands controller.", "error", err.Error())
	} else {
		subscriptions = append(subscriptions, commandController.Subscriptions()...)
		configs = append(configs, commandController.Configs()...)
	}

	// Create a new connection to the MQTT broker. This will also publish the
	// device subscriptions.
	client, err := mqttapi.NewClient(mqttCtx, agent.prefs, subscriptions, configs)
	if err != nil {
		logging.FromContext(ctx).Error("Could not connect to MQTT.", "error", err.Error())

		return
	}

	go func() {
		logging.FromContext(ctx).Debug("Listening for messages to publish to MQTT.")

		for {
			select {
			case msg := <-deviceController.Msgs():
				if err := client.Publish(msg); err != nil {
					logging.FromContext(ctx).Warn("Unable to publish message to MQTT.", "topic", msg.Topic, "content", slog.Any("msg", msg.Message))
				}
			case <-ctx.Done():
				mqttCancel()
				logging.FromContext(ctx).Debug("Stopped listening for messages to publish to MQTT.")

				return
			}
		}
	}()

	<-ctx.Done()
}

func (agent *Agent) resetMQTTWorker(ctx context.Context) error {
	if !agent.prefs.MQTTEnabled {
		return nil
	}

	mqttDevice := newMQTTDevice(ctx)

	client, err := mqttapi.NewClient(ctx, agent.prefs, nil, nil)
	if err != nil {
		return fmt.Errorf("could not connect to MQTT: %w", err)
	}

	if err := client.Unpublish(mqttDevice.Configs()...); err != nil {
		return fmt.Errorf("could not remove configs from MQTT: %w", err)
	}

	return nil
}
