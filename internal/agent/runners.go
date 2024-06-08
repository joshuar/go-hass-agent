// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/internal/scripts"
)

// Worker represents an object that is responsible for controlling the
// publishing of one or more sensors.
type Worker interface {
	// Name is the collective name of the sensors this worker controls.
	Name() string
	// Description is a longer text of what particular sensors are gathered and
	// where from.
	Description() string
	// Sensors returns an array of the current value of all sensors, or a
	// non-nil error if this is not possible.
	Sensors(ctx context.Context) ([]sensor.Details, error)
	// Updates returns a channel on which updates to sensors will be published,
	// when they become available.
	Updates(ctx context.Context) (<-chan sensor.Details, error)
}

// runWorkers will call all the sensor worker functions that have been defined
// for this device.
func runWorkers(ctx context.Context, trk SensorTracker, reg sensor.Registry) {
	workers := sensorWorkers()
	workers = append(workers, device.NewExternalIPUpdaterWorker(), device.NewVersionWorker())

	outCh := make([]<-chan sensor.Details, 0, len(workers))

	cancelFuncs := make([]context.CancelFunc, 0, len(workers))

	log.Debug().Msg("Starting worker funcs.")
	for worker := range len(workers) {
		workerCtx, cancelFunc := context.WithCancel(ctx)
		cancelFuncs = append(cancelFuncs, cancelFunc)

		log.Debug().Str("name", workers[worker].Name()).Str("description", workers[worker].Description()).Msg("Starting sensor worker.")

		workerCh, err := workers[worker].Updates(workerCtx)
		if err != nil {
			log.Warn().Err(err).Str("name", workers[worker].Name()).Msg("Could not start worker.")

			continue
		}

		outCh = append(outCh, workerCh)
	}

	sensorUpdates := sensor.MergeSensorCh(ctx, outCh...)
	go func() {
		log.Debug().Msg("Listening for sensor updates.")
		for update := range sensorUpdates {
			go func(update sensor.Details) {
				if err := trk.UpdateSensor(ctx, reg, update); err != nil {
					log.Warn().Err(err).Str("id", update.ID()).Msg("Update failed.")
				} else {
					log.Debug().
						Str("name", update.Name()).
						Str("id", update.ID()).
						Interface("state", update.State()).
						Str("units", update.Units()).
						Msg("Sensor updated.")
				}
			}(update)
		}
	}()

	go func() {
		<-ctx.Done()
		for _, c := range cancelFuncs {
			c()
		}
	}()
}

// runScripts will retrieve all scripts that the agent can run and queue them up
// to be run on their defined schedule using the cron scheduler. It also sets up
// a channel to receive script output and send appropriate sensor objects to the
// sensor.
func runScripts(ctx context.Context, path string, trk SensorTracker, reg sensor.Registry) {
	allScripts, err := scripts.FindScripts(path)

	switch {
	case err != nil:
		log.Error().Err(err).Msg("Error getting scripts.")
		return
	case len(allScripts) == 0:
		log.Debug().Msg("Could not find any script files.")
		return
	}

	scheduler := cron.New()

	outCh := make([]<-chan sensor.Details, 0, len(allScripts))

	for _, script := range allScripts {
		schedule := script.Schedule()
		if schedule != "" {
			_, err := scheduler.AddJob(schedule, script)
			if err != nil {
				log.Warn().Err(err).Str("script", script.Path()).
					Msg("Unable to schedule script.")

				break
			}

			outCh = append(outCh, script.Output)
			log.Debug().Str("schedule", schedule).Str("script", script.Path()).
				Msg("Added script sensor.")
		}
	}
	log.Debug().Msg("Starting cron scheduler for script sensors.")
	scheduler.Start()
	go func() {
		for scriptUpdates := range sensor.MergeSensorCh(ctx, outCh...) {
			go func(update sensor.Details) {
				if err := trk.UpdateSensor(ctx, reg, update); err != nil {
					log.Warn().Err(err).Str("id", update.ID()).Msg("Update sensor failed.")
				} else {
					log.Debug().
						Str("name", update.Name()).
						Str("id", update.ID()).
						Interface("state", update.State()).
						Str("units", update.Units()).
						Msg("Sensor updated.")
				}
			}(scriptUpdates)
		}
	}()
	<-ctx.Done()
	log.Debug().Msg("Stopping cron scheduler for script sensors.")
	cronCtx := scheduler.Stop()
	<-cronCtx.Done()
}

// runNotificationsWorker will run a goroutine that is listening for
// notification messages from Home Assistant on a websocket connection. Any
// received notifications will be dipslayed on the device running the agent.
func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	log.Debug().Msg("Listening for notifications.")

	notifyCh := hass.StartWebsocket(ctx)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopping notification handler.")
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
func runMQTTWorker(ctx context.Context) {
	prefs, err := preferences.Load()
	if err != nil {
		log.Error().Err(err).Msg("Could not load MQTT preferences.")
		return
	}
	if !prefs.MQTTEnabled {
		return
	}

	mqttCtx, mqttCancel := context.WithCancel(ctx)
	defer mqttCancel()

	// Create an MQTT device for this operating system and run its Setup.
	mqttDevice := newMQTTDevice(mqttCtx)
	if err = mqttDevice.Setup(mqttCtx); err != nil {
		log.Error().Err(err).Msg("Could not set up device MQTT functionality.")
		return
	}

	// Create a new connection to the MQTT broker. This will also publish the
	// device subscriptions.
	client, err := mqttapi.NewClient(mqttCtx, prefs, mqttDevice.Subscriptions(), mqttDevice.Configs())
	if err != nil {
		log.Error().Err(err).Msg("Could not connect to MQTT broker.")
		return
	}

	// Publish the device configs.
	log.Debug().Msg("Publishing configs.")
	if err := client.Publish(mqttDevice.Configs()...); err != nil {
		log.Error().Err(err).Msg("Failed to publish configuration messages.")
	}

	go func() {
		log.Debug().Msg("Listening for messages to publish to MQTT.")

		for {
			select {
			case msg := <-mqttDevice.Msgs():
				if err := client.Publish(msg); err != nil {
					log.Warn().Err(err).Msg("Unable to publish message to MQTT.")
				}
			case <-ctx.Done():
				mqttCancel()
				log.Debug().Msg("Stopped listening for messages to publish to MQTT.")
				return
			}
		}
	}()

	<-ctx.Done()
}

func resetMQTTWorker(ctx context.Context) {
	prefs, err := preferences.Load()
	if err != nil {
		log.Error().Err(err).Msg("Could not load MQTT preferences.")
		return
	}
	if !prefs.MQTTEnabled {
		return
	}

	mqttDevice := newMQTTDevice(ctx)

	client, err := mqttapi.NewClient(ctx, prefs, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("Could not connect to MQTT broker.")
		return
	}

	if err := client.Unpublish(mqttDevice.Configs()...); err != nil {
		log.Error().Err(err).Msg("Failed to reset MQTT.")
	}
}
