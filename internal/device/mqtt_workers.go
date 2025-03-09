// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"
	"errors"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/mqtt/commands"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

func CreateDeviceMQTTWorkers(ctx context.Context) []workers.MQTTWorker {
	var mqttWorkers []workers.MQTTWorker
	// Set up custom MQTT commands worker.
	customCommandsWorker, err := commands.NewCommandsWorker(ctx, preferences.MQTTDevice())
	if err != nil {
		if !errors.Is(err, commands.ErrNoCommands) {
			logging.FromContext(ctx).Warn("Could not setup custom MQTT commands.",
				slog.Any("error", err))
		}
	} else {
		mqttWorkers = append(mqttWorkers, customCommandsWorker)
	}

	return mqttWorkers
}
