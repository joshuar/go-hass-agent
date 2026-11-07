// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/eclipse/paho.golang/paho"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
)

const (
	// haStatusOnline is the payload Home Assistant publishes on its status
	// topic when it comes online.
	haStatusOnline = "online"
	// stateRepublishDelay is how long to wait after Home Assistant comes online
	// before republishing entity states. Home Assistant first needs to process
	// the entity configs (republished by the MQTT client), otherwise the states
	// of entities it has not yet created are dropped.
	stateRepublishDelay = 5 * time.Second
)

// WorkerData contains the configs, subscriptions and message channels for an
// MQTT worker or workers.
type WorkerData struct {
	Configs       []*models.MQTTConfig
	Subscriptions []*models.MQTTSubscription
	Msgs          <-chan models.MQTTMsg
	// States, if not nil, returns the current state messages of the worker
	// entities. These are republished whenever Home Assistant comes online.
	States func() []*models.MQTTMsg
}

// Start will connect to MQTT, publish worker configs and subscriptions, then
// start a goroutine to listen for messages from workers to publish through the
// client. If the client connection fails, a non-nil error is returned.
func Start(ctx context.Context, data *WorkerData) error {
	// Load the mqtt config.
	var mqttcfg Config
	if err := config.Load(ConfigPrefix, &mqttcfg); err != nil {
		return fmt.Errorf("unable to start MQTT: %w", err)
	}
	if ok, err := mqttcfg.Valid(); err != nil || !ok {
		return fmt.Errorf("load mqtt config: %w", err)
	}
	// When Home Assistant (re)starts, it recreates the entities from their
	// configs but has no state for them until a worker next publishes one,
	// which might not happen for a long time. Listen for Home Assistant coming
	// online so that the current states can be republished.
	haOnline := make(chan struct{}, 1)
	if data.States != nil {
		data.Subscriptions = append(data.Subscriptions, &models.MQTTSubscription{
			Topic: mqttcfg.TopicPrefix() + "/status",
			Callback: func(p *paho.Publish) {
				if string(p.Payload) != haStatusOnline {
					return
				}
				// Don't block the MQTT client; a pending republish covers
				// this one too.
				select {
				case haOnline <- struct{}{}:
				default:
				}
			},
		})
	}
	// Create a new connection to the MQTT broker, publish subscriptions and
	// configs.
	client, err := mqttapi.NewClient(ctx, &mqttcfg, data.Subscriptions, data.Configs)
	if err != nil {
		return fmt.Errorf("could not start MQTT client: %w", err)
	}
	if data.States != nil {
		go republishStates(ctx, haOnline, stateRepublishDelay, data.States, client.Publish)
	}
	// Listen for worker MQTT messages and publish them through the client.
	go func() {
		for {
			select {
			case msg := <-data.Msgs:
				if err := client.Publish(ctx, &msg); err != nil {
					slogctx.FromCtx(ctx).Warn("Unable to publish message to MQTT.",
						slog.String("topic", msg.Topic),
						slog.Any("msg", msg.Message))
				}
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Stopped listening for messages to publish to MQTT.")
				return
			}
		}
	}()

	return nil
}

// republishStates publishes the current entity states each time Home Assistant
// signals it has come online, after waiting the given delay.
func republishStates(
	ctx context.Context,
	haOnline <-chan struct{},
	delay time.Duration,
	states func() []*models.MQTTMsg,
	publish func(context.Context, ...*models.MQTTMsg) error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-haOnline:
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			// Any online signal received while waiting is covered by this
			// republish.
			select {
			case <-haOnline:
			default:
			}
			msgs := states()
			slogctx.FromCtx(ctx).Debug("Home Assistant online, republishing entity states.",
				slog.Int("count", len(msgs)))
			if err := publish(ctx, msgs...); err != nil {
				slogctx.FromCtx(ctx).Warn("Unable to republish entity states to MQTT.",
					slog.Any("error", err))
			}
		}
	}
}

// Reset will connect to MQTT and unpublish worker configs. If there is an
// problem, a non-nil error is returned.
func Reset(ctx context.Context, configs []*models.MQTTConfig) error {
	// Load the mqtt config.
	var mqttcfg Config
	if err := config.Load(ConfigPrefix, &mqttcfg); err != nil {
		return fmt.Errorf("could not reset MQTT preferences: %w", err)
	}
	client, err := mqttapi.NewClient(ctx, &mqttcfg, nil, nil)
	if err != nil {
		return fmt.Errorf("could not reset MQTT preferences: %w", err)
	}
	if err := client.Unpublish(ctx, configs...); err != nil {
		return fmt.Errorf("could not reset MQTT preferences: %w", err)
	}

	return nil
}
