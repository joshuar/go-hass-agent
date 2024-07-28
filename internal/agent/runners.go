// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate moq -out runners_mocks_test.go . SensorController Worker Script
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/robfig/cron/v3"

	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/scripts"
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

// runWorkers will call all the sensor worker functions that have been defined
// for this device.
func (agent *Agent) runWorkers(ctx context.Context, trk SensorTracker, reg sensor.Registry, controllers ...SensorController) {
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

		return
	}

	if err := trk.Process(ctx, reg, sensorCh...); err != nil {
		agent.logger.Error("Could not process sensor updates", "error", err.Error())
	}

	for _, controller := range controllers {
		if err := controller.StopAll(); err != nil {
			agent.logger.Warn("Stop controller had errors.", "error", err.Error())
		}
	}
}

// runScripts will retrieve all scripts that the agent can run and queue them up
// to be run on their defined schedule using the cron scheduler. It also sets up
// a channel to receive script output and send appropriate sensor objects to the
// sensor.
func (agent *Agent) runScripts(ctx context.Context, trk SensorTracker, reg sensor.Registry, sensorScripts ...Script) {
	if len(sensorScripts) == 0 {
		agent.logger.Warn("No sensor scripts to run.")

		return
	}

	scheduler := cron.New()
	sensorCh := make(chan sensor.Details)

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
	// Process any sensors returned by scripts as they are executed by the
	// scheduler.
	if err := trk.Process(ctx, reg, sensorCh); err != nil {
		agent.logger.Error("Could not process script sensor updates", "error", err.Error())
	}

	go func() {
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
}

// runNotificationsWorker will run a goroutine that is listening for
// notification messages from Home Assistant on a websocket connection. Any
// received notifications will be dipslayed on the device running the agent.
func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	notifyCh, err := hass.StartWebsocket(ctx)
	if err != nil {
		agent.logger.Error("Could not listen for notifications.", "error", err.Error())
	}

	agent.logger.Debug("Listening for notifications.")

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				agent.logger.Debug("Stopping notification handler.")

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
func (agent *Agent) runMQTTWorker(ctx context.Context, osController MQTTController, commandsFile string) {
	var (
		commandController MQTTController
		subscriptions     []*mqttapi.Subscription
		configs           []*mqttapi.Msg
		err               error
	)

	// Create an MQTT device for this operating system and run its Setup.
	subscriptions = append(subscriptions, osController.Subscriptions()...)
	configs = append(configs, osController.Configs()...)

	// Create an MQTT device for this operating system and run its Setup.
	commandController, err = commands.NewCommandsController(ctx, commandsFile, device.MQTTDeviceInfo(ctx))
	if err != nil {
		agent.logger.Warn("Could not set up MQTT commands controller.", "error", err.Error())
	} else {
		subscriptions = append(subscriptions, commandController.Subscriptions()...)
		configs = append(configs, commandController.Configs()...)
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
			case msg := <-osController.Msgs():
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

// FindScripts locates scripts and returns a slice of scripts that the agent can
// run.
func findScripts(path string) ([]Script, error) {
	var sensorScripts []Script

	var errs error

	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not search for scripts: %w", err)
	}

	for _, scriptFile := range files {
		if isExecutable(scriptFile) {
			script, err := scripts.NewScript(scriptFile)
			if err != nil {
				errs = errors.Join(errs, err)

				continue
			}

			sensorScripts = append(sensorScripts, script)
		}
	}

	return sensorScripts, nil
}

func isExecutable(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.Mode().Perm()&0o111 != 0
}
