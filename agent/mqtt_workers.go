// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt/commands"
)

func CreateDeviceMQTTWorkers(ctx context.Context) ([]workers.MQTTWorker, error) {
	var mqttWorkers []workers.MQTTWorker
	// Set up custom MQTT commands worker.
	device, err := mqtt.MQTTDevice()
	if err != nil {
		return nil, fmt.Errorf("unable to create device MQTT workers: %w", err)
	}
	customCommandsWorker, err := commands.NewCommandsWorker(ctx, device)
	if err != nil {
		if !errors.Is(err, commands.ErrNoCommands) {
			slogctx.FromCtx(ctx).Warn("Could not setup custom MQTT commands.",
				slog.Any("error", err))
		}
	} else {
		mqttWorkers = append(mqttWorkers, customCommandsWorker)
	}

	return mqttWorkers, nil
}
