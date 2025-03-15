// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"errors"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/mqtt"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

var ErrReset = errors.New("error resetting agent")

// Reset is invoked when Go Hass Agent is run with the `reset` command-line
// option (i.e., `go-hass-agent reset`).
func Reset(ctx context.Context) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	// Add device-specific values to context.
	ctx = device.SetupCtx(ctx)

	manager := workers.NewManager(ctx)

	// Create MQTT workers.
	var mqttWorkers []workers.MQTTWorker
	// Add device-based MQTT workers.
	mqttWorkers = append(mqttWorkers, device.CreateDeviceMQTTWorkers(ctx)...)
	// Add os-based MQTT workers.
	mqttWorkers = append(mqttWorkers, device.CreateOSMQTTWorkers(ctx))
	data := manager.StartMQTTWorkers(ctx, mqttWorkers...)

	if err := mqtt.Reset(ctx, data.Configs); err != nil {
		return errors.Join(ErrReset, err)
	}

	return nil
}
