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
	"sync"

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/preferences"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"
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
		wg            sync.WaitGroup
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

	wg.Add(1)

	go func() {
		defer wg.Done()
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

	wg.Wait()
}

func (agent *Agent) resetMQTTControllers(ctx context.Context) error {
	mqttDevice := agent.newMQTTDevice()

	var configs []*mqttapi.Msg

	_, osMQTTController := agent.newOSController(ctx, mqttDevice)
	if osMQTTController != nil {
		configs = append(configs, osMQTTController.Configs()...)
	}

	mqttCmdController := agent.newMQTTController(ctx, mqttDevice)
	if mqttCmdController != nil {
		configs = append(configs, mqttCmdController.Configs()...)
	}

	client, err := mqttapi.NewClient(ctx, agent.prefs.GetMQTTPreferences(), nil, nil)
	if err != nil {
		return fmt.Errorf("could not connect to MQTT: %w", err)
	}

	if err := client.Unpublish(ctx, configs...); err != nil {
		return fmt.Errorf("could not remove configs from MQTT: %w", err)
	}

	return nil
}

func (agent *Agent) newMQTTDevice() *mqtthass.Device {
	mqttDevice, err := device.MQTTDevice(agent.prefs.Device.Name, agent.prefs.Device.ID, preferences.AppURL, preferences.AppVersion)
	if err != nil {
		agent.logger.Warn("Could not create MQTT device, MQTT functionality will not be available.", slog.Any("error", err))
	}

	return mqttDevice
}
