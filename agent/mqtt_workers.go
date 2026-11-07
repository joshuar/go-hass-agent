// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt/commands"
)

// CreateDeviceMQTTWorkers sets up the device-specific MQTT workers.
func CreateDeviceMQTTWorkers() ([]workers.MQTTWorker, error) {
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

// stateMarshaller is satisfied by any entity whose current state can be
// marshalled into a message.
type stateMarshaller interface {
	MarshalState(args ...any) (*mqttapi.Msg, error)
}

// appendStates appends the current state message of each of the given entities
// to msgs. Entities without a state callback are skipped.
func appendStates[T stateMarshaller](logger *slog.Logger, msgs []*mqttapi.Msg, entities []T) []*mqttapi.Msg {
	for _, entity := range entities {
		msg, err := entity.MarshalState()
		if err != nil {
			if !errors.Is(err, mqtthass.ErrNoStateCallback) {
				logger.Warn("Could not create state.", slog.Any("error", err))
			}

			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs
}
