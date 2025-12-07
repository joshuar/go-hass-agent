// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt/commands"
)

// CreateDeviceMQTTWorkers sets up the device-specific MQTT workers.
func CreateDeviceMQTTWorkers(ctx context.Context) ([]workers.MQTTWorker, error) {
	var mqttWorkers []workers.MQTTWorker
	// Set up custom MQTT commands worker.
	device, err := mqtt.Device()
	if err != nil {
		return nil, fmt.Errorf("create mqtt device: %w", err)
	}
	customCommandsWorker, err := commands.NewCommandsWorker(device)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("create mqtt custom commands worker: %w", err)
		}
	} else {
		mqttWorkers = append(mqttWorkers, customCommandsWorker)
	}

	return mqttWorkers, nil
}
