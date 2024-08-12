// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

// powerController contains items that all power controls use.
type powerController struct {
	logger      *slog.Logger
	bus         *dbusx.Bus
	sessionPath string
}

// powerCommand represents a power command in the agent.
type powerCommand struct {
	callBack func(p *paho.Publish)
	name     string
	id       string
	icon     string
}

// generateCommands generates a list of power commands that are available on the
// device running the agent.
func (c *powerController) generateCommands(ctx context.Context) []powerCommand {
	systemCommands := []powerCommand{
		{
			name: "Reboot",
			id:   "Reboot",
			icon: "mdi:restart",
		},
		{
			name: "Suspend",
			id:   "Suspend",
			icon: "mdi:power-sleep",
		},
		{
			name: "Hibernate",
			id:   "Hibernate",
			icon: "mdi:power-sleep",
		},
		{
			name: "Power Off",
			icon: "mdi:power",
			id:   "PowerOff",
		},
	}

	sessionCommands := []powerCommand{
		{
			name: "Lock Session",
			id:   "Lock",
			icon: "mdi:eye-lock",
		},
		{
			name: "Unlock Session",
			icon: "mdi:eye-lock-open",
			id:   "UnLock",
		},
	}

	availableCommands := make([]powerCommand, 0, len(systemCommands)+len(sessionCommands))

	// Add available system power commands.
	for _, config := range systemCommands {
		// Check if this power method is available on the system.
		available, _ := dbusx.GetData[string](ctx, c.bus, loginBasePath, loginBaseInterface, managerInterface+".Can"+config.id) //nolint:errcheck
		if available == "yes" {
			config.callBack = c.generatePowerControlCallback(ctx, config.name, loginBasePath, managerInterface+"."+config.id, true)
			availableCommands = append(availableCommands, config)
		}
	}

	// Add session power commands.
	for _, config := range sessionCommands {
		config.callBack = c.generatePowerControlCallback(ctx, config.name, c.sessionPath, sessionInterface+"."+config.id, nil)
		availableCommands = append(availableCommands, config)
	}

	return availableCommands
}

// generatePowerControlCallback creates an MQTT callback function that can
// execute the appropriate D-Bus call to issue a power command on the device.
func (c *powerController) generatePowerControlCallback(ctx context.Context, name, path, method string, arg any) func(p *paho.Publish) {
	if arg != nil {
		return func(_ *paho.Publish) {
			if err := c.bus.Call(ctx, path, loginBaseInterface, method, arg); err != nil {
				c.logger.Warn("Could not issue power control.", slog.String("control", name), slog.Any("error", err))
			}
		}
	}

	return func(_ *paho.Publish) {
		if err := c.bus.Call(ctx, path, loginBaseInterface, method); err != nil {
			c.logger.Warn("Could not issue power control.", slog.String("control", name), slog.Any("error", err))
		}
	}
}

//nolint:lll
func NewPowerControl(ctx context.Context, api *dbusx.DBusAPI, parentLogger *slog.Logger, device *mqtthass.Device) ([]*mqtthass.ButtonEntity, error) {
	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("cannot create power controls: %w", err)
	}

	sessionPath, err := bus.GetSessionPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create power controls: %w", err)
	}

	controller := &powerController{
		logger:      parentLogger.WithGroup("power_control"),
		bus:         bus,
		sessionPath: sessionPath,
	}

	var entities []*mqtthass.ButtonEntity //nolint:prealloc

	for _, command := range controller.generateCommands(ctx) {
		entities = append(entities,
			mqtthass.AsButton(
				mqtthass.NewEntity(preferences.AppName, command.name, command.name).
					WithOriginInfo(preferences.MQTTOrigin()).
					WithDeviceInfo(device).
					WithIcon(command.icon).
					WithCommandCallback(command.callBack)))
	}

	return entities, nil
}
