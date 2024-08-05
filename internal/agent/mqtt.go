// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func (agent *Agent) newMQTTCommandsController(ctx context.Context) MQTTController {
	commandsFile := filepath.Join(xdg.ConfigHome, agent.id, "commands.toml")

	mqttDeviceInfo, err := device.MQTTDevice(preferences.AppName, agent.id, preferences.AppURL, preferences.AppVersion)
	if err != nil {
		agent.logger.Warn("Could not set up MQTT commands controller.", "error", err.Error())

		return nil
	}

	commandController, err := commands.NewCommandsController(ctx, commandsFile, mqttDeviceInfo)
	if err != nil {
		agent.logger.Warn("Could not set up MQTT commands controller.", "error", err.Error())

		return nil
	}

	return commandController
}
