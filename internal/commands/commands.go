// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:dupl,exhaustruct
//revive:disable:unused-receiver
package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	"github.com/iancoleman/strcase"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// ErrNoCommands indicates there were no commands to configure.
var (
	ErrNoCommands         = errors.New("no commands")
	ErrUnknownSwitchState = errors.New("could not parse current state of switch")
)

// Command represents a Command to run by a button or switch.
type Command struct {
	// Name is display name for the command.
	Name string `toml:"name"`
	// Exec is the actual binary or script to run.
	Exec string `toml:"exec"`
	// Icon is a material design icon representing the command.
	Icon string `toml:"icon,omitempty"`
}

// CommandList is a CommandList of all the buttons/commands parsed from the config file.
//
//nolint:tagalign
//revive:disable:struct-tag
type CommandList struct {
	Buttons  []Command `toml:"button,omitempty" koanf:"button"`
	Switches []Command `toml:"switch,omitempty" koanf:"switch"`
}

// Controller represents an object with one or more buttons and switches
// definitions, which can be passed to Home Assistant to add appropriate
// entities to control the buttons/switches over MQTT.
type Controller struct {
	buttons  []*mqtthass.ButtonEntity
	switches []*mqtthass.SwitchEntity
}

// Subscriptions are the MQTT subscriptions for buttons and switches, providing
// the appropriate callback mechanism to execute the associated commands.
func (d *Controller) Subscriptions() []*mqttapi.Subscription {
	var subs []*mqttapi.Subscription

	// Create subscriptions for buttons.
	if d.buttons != nil {
		for _, button := range d.buttons {
			if sub, err := button.MarshalSubscription(); err != nil {
				log.Warn().Err(err).Str("entity", button.Name).Msg("Could not create subscription.")
			} else {
				subs = append(subs, sub)
			}
		}
	}
	// Create subscriptions for switches.
	if d.switches != nil {
		for _, sw := range d.switches {
			if sub, err := sw.MarshalSubscription(); err != nil {
				log.Warn().Err(err).Str("entity", sw.Name).Msg("Could not create subscription.")
			} else {
				subs = append(subs, sub)
			}
		}
	}

	return subs
}

// Configs are the MQTT configurations required by Home Assistant to set up
// entities for the buttons/switches.
func (d *Controller) Configs() []*mqttapi.Msg {
	var configs []*mqttapi.Msg

	// Create button configs.
	if d.buttons != nil {
		for _, button := range d.buttons {
			if sub, err := button.MarshalConfig(); err != nil {
				log.Warn().Err(err).Str("entity", button.Name).Msg("Could not create subscription.")
			} else {
				configs = append(configs, sub)
			}
		}
	}
	// Create switch configs.
	if d.switches != nil {
		for _, sw := range d.switches {
			if sub, err := sw.MarshalConfig(); err != nil {
				log.Warn().Err(err).Str("entity", sw.Name).Msg("Could not create subscription.")
			} else {
				configs = append(configs, sub)
			}
		}
	}

	return configs
}

// Msgs are additional MQTT messages to be published based on any event logic
// managed by the controller. This is unused.
func (d *Controller) Msgs() chan *mqttapi.Msg {
	return nil
}

// Setup can be used to initialise the controller. This is unused.
func (d *Controller) Setup(_ context.Context) error {
	return nil
}

// NewCommandsController is used by the agent to initialise the commands
// controller, which holds the MQTT configuration for the commands defined by
// the user.
//
//nolint:exhaustruct
func NewCommandsController(ctx context.Context, commandsFile string, device *mqtthass.Device) (*Controller, error) {
	if _, err := os.Stat(commandsFile); errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoCommands
	}

	data, err := os.ReadFile(commandsFile)
	if err != nil {
		return nil, fmt.Errorf("could not read commands file: %w", err)
	}

	cmds := &CommandList{}

	if err := toml.Unmarshal(data, &cmds); err != nil {
		return nil, fmt.Errorf("could not parse commands file: %w", err)
	}

	controller := newController(ctx, device, cmds)

	return controller, nil
}

// newController creates a new MQTT controller to manage a bunch of buttons and
// switches a user has defined.
func newController(_ context.Context, device *mqtthass.Device, commands *CommandList) *Controller {
	controller := &Controller{
		buttons:  generateButtons(commands.Buttons, device),
		switches: generateSwitches(commands.Switches, device),
	}

	return controller
}

// generateButtons will create MQTT entities for buttons defined by the
// controller.
func generateButtons(buttonCmds []Command, device *mqtthass.Device) []*mqtthass.ButtonEntity {
	var id, icon, name string

	entities := make([]*mqtthass.ButtonEntity, 0, len(buttonCmds))

	for _, cmd := range buttonCmds {
		callback := func(_ *paho.Publish) {
			err := buttonCmd(cmd.Exec)
			if err != nil {
				log.Warn().Err(err).Str("command", cmd.Name).Msg("Button press failed.")
			}
		}
		name = cmd.Name
		id = strcase.ToSnake(device.Name + "_" + cmd.Name)

		if cmd.Icon != "" {
			icon = cmd.Icon
		} else {
			icon = "mdi:button-pointer"
		}

		entities = append(entities,
			mqtthass.AsButton(
				mqtthass.NewEntity(preferences.AppName, name, id).
					WithOriginInfo(preferences.MQTTOrigin()).
					WithDeviceInfo(device).
					WithIcon(icon).
					WithCommandCallback(callback)))
	}

	return entities
}

// buttonCmd runs the executable associated with a button. Buttons are not
// expected to accept any input, or produce any consumable output, so only the
// return value is checked.
func buttonCmd(command string) error {
	cmdElems := strings.Split(command, " ")

	_, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return fmt.Errorf("could not execute button command: %w", err)
	}

	return nil
}

// generateButtons will create MQTT entities for buttons defined by the
// controller.
func generateSwitches(switchCmds []Command, device *mqtthass.Device) []*mqtthass.SwitchEntity {
	var id, icon, name string

	entities := make([]*mqtthass.SwitchEntity, 0, len(switchCmds))

	for _, cmd := range switchCmds {
		cmdCallBack := func(p *paho.Publish) {
			state := string(p.Payload)

			err := switchCmd(cmd.Exec, state)
			if err != nil {
				log.Warn().Err(err).Str("command", cmd.Name).Msg("Switch change failed.")
			}
		}
		stateCallBack := func(_ ...any) (json.RawMessage, error) {
			return switchState(cmd.Exec)
		}
		name = cmd.Name
		id = strcase.ToSnake(device.Name + "_" + cmd.Name)

		if cmd.Icon != "" {
			icon = cmd.Icon
		} else {
			icon = "mdi:toggle-switch"
		}

		entities = append(entities,
			mqtthass.AsSwitch(
				mqtthass.NewEntity(preferences.AppName, name, id).
					WithOriginInfo(preferences.MQTTOrigin()).
					WithDeviceInfo(device).
					WithIcon(icon).
					WithStateCallback(stateCallBack).
					WithCommandCallback(cmdCallBack),
				true))
	}

	return entities
}

// buttonCmd runs the executable associated with a button. Buttons are not
// expected to accept any input, or produce any consumable output, so only the
// return value is checked.
func switchCmd(command, state string) error {
	cmdElems := strings.Split(command, " ")
	cmdElems = append(cmdElems, state)

	_, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return fmt.Errorf("could not execute button command: %w", err)
	}

	return nil
}

func switchState(command string) (json.RawMessage, error) {
	cmdElems := strings.Split(command, " ")

	output, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return nil, fmt.Errorf("could get switch state: %w", err)
	}

	switch {
	case bytes.Contains(output, []byte(`ON`)):
		return json.RawMessage(`ON`), nil
	case bytes.Contains(output, []byte(`OFF`)):
		return json.RawMessage(`OFF`), nil
	}

	return nil, ErrUnknownSwitchState
}
