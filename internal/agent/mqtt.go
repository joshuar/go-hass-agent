// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// MQTTWorker represents an object that is responsible for controlling the
// publishing of data over MQTT.
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

// setupMQTT will load a context with MQTT preferences and device configuration.
func setupMQTT(ctx context.Context) (context.Context, error) {
	// Get the MQTT preferences.
	prefs, err := preferences.GetMQTTPreferences()
	if err != nil {
		return ctx, fmt.Errorf("could not get MQTT preferences: %w", err)
	}
	// Add MQTT preferences to context.
	ctx = MQTTPrefsToCtx(ctx, prefs)
	// Get MQTT device and add to context.
	ctx = MQTTDeviceToCtx(ctx, preferences.GetMQTTDevice())

	return ctx, nil
}

// createMQTTWorkers creates the MQTT workers.
func createMQTTWorkers(ctx context.Context) []MQTTWorker {
	var workers []MQTTWorker
	// Set up custom MQTT commands worker.
	customCommandsWorker, err := commands.NewCommandsWorker(ctx, MQTTDeviceFromFromCtx(ctx))
	if err != nil {
		if !errors.Is(err, commands.ErrNoCommands) {
			logging.FromContext(ctx).Warn("Could not setup custom MQTT commands.",
				slog.Any("error", err))
		}
	} else {
		workers = append(workers, customCommandsWorker)
	}
	// Set up OS MQTT worker.
	workers = append(workers, setupOSMQTTWorker(ctx))

	return workers
}

// processMQTTWorkers will connect to MQTT, publish configs and subscriptions and
// listen for any messages from all MQTT workers.
func processMQTTWorkers(ctx context.Context) {
	var ( //nolint:prealloc
		subscriptions []*mqttapi.Subscription
		configs       []*mqttapi.Msg
		msgCh         []<-chan *mqttapi.Msg
		err           error
	)

	if !preferences.MQTTEnabled() {
		return
	}

	// Get the MQTT preferences and device.
	ctx, err = setupMQTT(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Could not set-up MQTT.",
			slog.Any("error", err))
		return
	}
	// Create the workers.
	workers := createMQTTWorkers(ctx)
	// Add the subscriptions and configs from the workers.
	for _, worker := range workers {
		subscriptions = append(subscriptions, worker.Subscriptions()...)
		configs = append(configs, worker.Configs()...)
		msgCh = append(msgCh, worker.Msgs())
	}
	// Create a new connection to the MQTT broker, publish subscriptions and
	// configs.
	client, err := mqttapi.NewClient(ctx, MQTTPrefsFromFromCtx(ctx), subscriptions, configs)
	if err != nil {
		logging.FromContext(ctx).Error("Could not connect to MQTT.",
			slog.Any("error", err))
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

// resetMQTTWorkers will unpublish configs for all defined MQTT workers.
func resetMQTTWorkers(ctx context.Context) error {
	var (
		configs []*mqttapi.Msg
		err     error
	)

	// Get the MQTT preferences and device.
	ctx, err = setupMQTT(ctx)
	if err != nil {
		return errors.New("could not reset MQTT: set-up failed")
	}
	// Create the workers.
	workers := createMQTTWorkers(ctx)

	for _, worker := range workers {
		configs = append(configs, worker.Configs()...)
	}

	client, err := mqttapi.NewClient(ctx, MQTTPrefsFromFromCtx(ctx), nil, nil)
	if err != nil {
		return fmt.Errorf("could not connect to MQTT: %w", err)
	}

	if err := client.Unpublish(ctx, configs...); err != nil {
		return fmt.Errorf("could not remove configs from MQTT: %w", err)
	}

	return nil
}
