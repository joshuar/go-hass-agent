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

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

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
}

// stateEntity is a convienience interface to avoid duplicating a lot of loop content
// when configuring the controller.
type stateEntity interface {
	MarshalConfig() (*mqttapi.Msg, error)
}

// commandEntity is a convienience interface to avoid duplicating a lot of loop content
// when configuring the controller.
type commandEntity interface {
	stateEntity
	MarshalSubscription() (*mqttapi.Subscription, error)
}

// mqttEntities holds all MQTT entities for a worker and a channel through which
// they can send MQTT messages.
type mqttEntities struct {
	msgs          chan *mqttapi.Msg
	sensors       []*mqtthass.SensorEntity
	buttons       []*mqtthass.ButtonEntity
	numbers       []*mqtthass.NumberEntity[int]
	switches      []*mqtthass.SwitchEntity
	controls      []*mqttapi.Subscription
	binarySensors []*mqtthass.SensorEntity
	cameras       []*mqtthass.CameraEntity
}

func setupMQTT(ctx context.Context, prefs *preferences.Preferences) []MQTTWorker {
	var workers []MQTTWorker

	// Create an MQTT device, used to configure MQTT functionality for some
	// controllers.
	mqttDevice := prefs.GenerateMQTTDevice(ctx)

	// Set up custom MQTT commands worker.
	customCommandsWorker, err := commands.NewCommandsWorker(ctx, mqttDevice)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not setup custom MQTT commands.",
			slog.Any("error", err))
	} else {
		workers = append(workers, customCommandsWorker)
	}

	osWorker := setupOSMQTTWorker(ctx, mqttDevice)
	workers = append(workers, osWorker)

	return workers
}

// processMQTTWorkers will connect to MQTT, publish configs and subscriptions and
// listen for any messages from all MQTT workers defined by the passed in
// MQTT controllers.
func processMQTTWorkers(ctx context.Context, prefs mqttapi.Preferences, controllers ...MQTTWorker) {
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
	client, err := mqttapi.NewClient(ctx, prefs, subscriptions, configs)
	if err != nil {
		logging.FromContext(ctx).Error("Could not connect to MQTT.", slog.Any("error", err))
		return
	}

	logging.FromContext(ctx).Debug("Listening for messages to publish to MQTT.")

	for {
		select {
		case msg := <-mergeCh(ctx, msgCh...):
			if err := client.Publish(ctx, msg); err != nil {
				logging.FromContext(ctx).Warn("Unable to publish message to MQTT.",
					slog.String("topic", msg.Topic),
					slog.Any("msg", msg.Message))
			}
		case <-ctx.Done():
			logging.FromContext(ctx).Debug("Stopped listening for messages to publish to MQTT.")
			return
		}
	}
}

func resetMQTTControllers(ctx context.Context, prefs *preferences.Preferences) error {
	var configs []*mqttapi.Msg

	workers := setupMQTT(ctx, prefs)
	for _, worker := range workers {
		configs = append(configs, worker.Configs()...)
	}

	client, err := mqttapi.NewClient(ctx, prefs.GetMQTTPreferences(), nil, nil)
	if err != nil {
		return fmt.Errorf("could not connect to MQTT: %w", err)
	}

	if err := client.Unpublish(ctx, configs...); err != nil {
		return fmt.Errorf("could not remove configs from MQTT: %w", err)
	}

	return nil
}
