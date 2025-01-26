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

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	powerControlPreferencesID = "power_controls"
)

type powerControlWorker struct {
	prefs *preferences.CommonWorkerPrefs
}

// powerCommand represents a power command in the agent.
type powerCommand struct {
	callBack func(p *paho.Publish)
	name     string
	id       string
	method   string
	icon     string
}

// generateCommands generates a list of power commands that are available on the
// device running the agent.
func generatePowerCommands(ctx context.Context, bus *dbusx.Bus, device string) []powerCommand {
	systemCommands := []powerCommand{
		{
			name:   "Reboot",
			id:     device + "_reboot",
			method: "Reboot",
			icon:   "mdi:restart",
		},
		{
			name:   "Suspend",
			id:     device + "_suspend",
			method: "Suspend",
			icon:   "mdi:power-sleep",
		},
		{
			name:   "Hibernate",
			id:     device + "_hibernate",
			method: "Hibernate",
			icon:   "mdi:power-sleep",
		},
		{
			name:   "Power Off",
			id:     device + "_poweroff",
			method: "PowerOff",
			icon:   "mdi:power",
		},
	}

	availableCommands := make([]powerCommand, 0, len(systemCommands))

	// Add available system power commands.
	for _, config := range systemCommands {
		// Check if this power method is available on the system.
		available, err := dbusx.GetData[string](bus, loginBasePath, loginBaseInterface, managerInterface+".Can"+config.method)
		if available == "yes" || err == nil {
			config.callBack = generatePowerControlCallback(ctx, bus, config.name, loginBasePath, managerInterface+"."+config.method)
			availableCommands = append(availableCommands, config)
		}
	}

	return availableCommands
}

// generatePowerControlCallback creates an MQTT callback function that can
// execute the appropriate D-Bus call to issue a power command on the device.
func generatePowerControlCallback(ctx context.Context, bus *dbusx.Bus, name, path, method string) func(p *paho.Publish) {
	return func(_ *paho.Publish) {
		err := dbusx.NewMethod(bus, loginBaseInterface, path, method).Call(ctx, true)
		if err != nil {
			logging.FromContext(ctx).With(slog.String("controller", "power")).
				Warn("Could not issue power control.",
					slog.String("control", name),
					slog.Any("error", err))
		}
	}
}

func (w *powerControlWorker) PreferencesID() string {
	return powerControlPreferencesID
}

func (w *powerControlWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewPowerControl(ctx context.Context, device *mqtthass.Device) ([]*mqtthass.ButtonEntity, error) {
	var err error

	worker := &powerControlWorker{}

	worker.prefs, err = preferences.LoadWorker(worker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	if worker.prefs.IsDisabled() {
		return nil, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	_, ok = linux.CtxGetSessionPath(ctx)
	if !ok {
		return nil, linux.ErrNoSessionPath
	}

	commands := generatePowerCommands(ctx, bus, device.Name)
	entities := make([]*mqtthass.ButtonEntity, 0, len(commands))

	for _, command := range commands {
		entities = append(entities,
			mqtthass.NewButtonEntity().
				WithDetails(
					mqtthass.App(preferences.AppName),
					mqtthass.Name(command.name),
					mqtthass.ID(command.id),
					mqtthass.DeviceInfo(device),
					mqtthass.Icon(command.icon),
				).
				WithCommand(mqtthass.CommandCallback(command.callBack)),
		)
	}

	return entities, nil
}
