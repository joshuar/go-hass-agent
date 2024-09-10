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
	"path/filepath"

	"github.com/adrg/xdg"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func (agent *Agent) newMQTTController(ctx context.Context, mqttDevice *mqtthass.Device) MQTTController {
	// Don't set up if no MQTT device has been passed in.
	if mqttDevice == nil {
		agent.logger.Warn("Not setting up MQTT controller, no MQTT device.")

		return nil
	}

	commandsFile := filepath.Join(xdg.ConfigHome, agent.id, "commands.toml")

	commandController, err := commands.NewCommandsController(ctx, commandsFile, mqttDevice)
	if err != nil {
		if !errors.Is(err, commands.ErrNoCommands) {
			agent.logger.Warn("Could not set up MQTT commands controller.", slog.Any("error", err))
		}

		return nil
	}

	return commandController
}

// runMQTTWorkers will connect to MQTT, publish configs and subscriptions and
// listen for any messages from all MQTT workers defined by the passed in
// MQTT controllers.
func (agent *Agent) runMQTTWorkers(ctx context.Context, controllers ...MQTTController) {
	var ( //nolint:prealloc
		subscriptions []*mqttapi.Subscription
		configs       []*mqttapi.Msg
		msgCh         []<-chan *mqttapi.Msg
		err           error
	)

	prefs := agent.prefs.GetMQTTPreferences()
	if prefs == nil {
		agent.logger.Debug("No MQTT preferences found, not running MQTT controller.")
		return
	}

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
		agent.logger.Error("Could not connect to MQTT.", slog.Any("error", err))
		return
	}

	agent.logger.Debug("Listening for messages to publish to MQTT.")

	for {
		select {
		case msg := <-mergeCh(ctx, msgCh...):
			if err := client.Publish(ctx, msg); err != nil {
				agent.logger.Warn("Unable to publish message to MQTT.",
					slog.String("topic", msg.Topic),
					slog.Any("msg", msg.Message))
			}
		case <-ctx.Done():
			agent.logger.Debug("Stopped listening for messages to publish to MQTT.")
			return
		}
	}
}

func (agent *Agent) resetMQTTControllers(ctx context.Context) error {
	mqttDevice := agent.newMQTTDevice()

	prefs := agent.prefs.GetMQTTPreferences()
	if prefs == nil {
		return nil
	}

	var configs []*mqttapi.Msg

	_, osMQTTController := agent.newOSController(ctx, mqttDevice)
	if osMQTTController != nil {
		configs = append(configs, osMQTTController.Configs()...)
	}

	mqttCmdController := agent.newMQTTController(ctx, mqttDevice)
	if mqttCmdController != nil {
		configs = append(configs, mqttCmdController.Configs()...)
	}

	client, err := mqttapi.NewClient(ctx, prefs, nil, nil)
	if err != nil {
		return fmt.Errorf("could not connect to MQTT: %w", err)
	}

	if err := client.Unpublish(ctx, configs...); err != nil {
		return fmt.Errorf("could not remove configs from MQTT: %w", err)
	}

	return nil
}

func (agent *Agent) newMQTTDevice() *mqtthass.Device {
	// Retrieve the hardware model and manufacturer.
	model, manufacturer, err := device.GetHWProductInfo()
	if err != nil {
		agent.logger.Warn("Error creating MQTT device.", slog.Any("error", err))
	}

	var deviceName, deviceID string

	if agent.prefs != nil {
		deviceName = agent.prefs.DeviceName()
		deviceID = agent.prefs.DeviceID()
	}

	return &mqtthass.Device{
		Name:         deviceName,
		URL:          preferences.AppURL,
		SWVersion:    preferences.AppVersion,
		Manufacturer: manufacturer,
		Model:        model,
		Identifiers:  []string{agent.id, deviceName, deviceID},
	}
}
