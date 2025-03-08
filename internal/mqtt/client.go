// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mqtt

import (
	"context"
	"errors"
	"log/slog"

	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var ErrClient = errors.New("MQTT client error")

// Start will connect to MQTT, publish worker configs and subscriptions, then
// start a goroutine to listen for messages from workers to publish through the
// client. If the client connection fails, a non-nil error is returned.
func Start(ctx context.Context, configs []*models.MQTTConfig, subscriptions []*models.MQTTSubscription, msgs <-chan models.MQTTMsg) error {
	// Create a new connection to the MQTT broker, publish subscriptions and
	// configs.
	client, err := mqttapi.NewClient(ctx, preferences.MQTT(), subscriptions, configs)
	if err != nil {
		return errors.Join(ErrClient, err)
	}
	// Listen for worker MQTT messages and publish them through the client.
	go func() {
		for {
			select {
			case msg := <-msgs:
				if err := client.Publish(ctx, &msg); err != nil {
					logging.FromContext(ctx).Warn("Unable to publish message to MQTT.",
						slog.String("topic", msg.Topic),
						slog.Any("msg", msg.Message))
				}
			case <-ctx.Done():
				logging.FromContext(ctx).Debug("Stopped listening for messages to publish to MQTT.")
				return
			}
		}
	}()

	return nil
}

// Reset will connect to MQTT and unpublish worker configs. If there is an
// problem, a non-nil error is returned.
func Reset(ctx context.Context, configs []*models.MQTTConfig) error {
	client, err := mqttapi.NewClient(ctx, preferences.MQTT(), nil, nil)
	if err != nil {
		return errors.Join(ErrClient, err)
	}

	if err := client.Unpublish(ctx, configs...); err != nil {
		return errors.Join(ErrClient, err)
	}

	return nil
}
