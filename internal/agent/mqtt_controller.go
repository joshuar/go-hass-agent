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
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func newMQTTController(ctx context.Context, mqttDevice *mqtthass.Device) MQTTController {
	// Don't set up if no MQTT device has been passed in.
	if mqttDevice == nil {
		logging.FromContext(ctx).Warn("Not setting up MQTT controller, no MQTT device.")

		return nil
	}

	appID := preferences.AppIDFromContext(ctx)

	commandsFile := filepath.Join(xdg.ConfigHome, appID, "commands.toml")

	commandController, err := commands.NewCommandsController(ctx, commandsFile, mqttDevice)
	if err != nil {
		if !errors.Is(err, commands.ErrNoCommands) {
			logging.FromContext(ctx).Warn("Could not set up MQTT commands controller.", slog.Any("error", err))
		}

		return nil
	}

	return commandController
}

// runMQTTWorkers will connect to MQTT, publish configs and subscriptions and
// listen for any messages from all MQTT workers defined by the passed in
// MQTT controllers.
func runMQTTWorkers(ctx context.Context, prefs mqttapi.Preferences, controllers ...MQTTController) {
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

func resetMQTTControllers(ctx context.Context, device *mqtthass.Device, prefs mqttapi.Preferences) error {
	var configs []*mqttapi.Msg

	osMQTTController := newOSMQTTController(ctx, device)
	configs = append(configs, osMQTTController.Configs()...)

	mqttCmdController := newMQTTController(ctx, device)
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
