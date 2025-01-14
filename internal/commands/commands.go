// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/constraints"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	stateValueTemplate = "{{ value_json.value }}"
	switchOnState      = "ON"
	switchOffState     = "OFF"
	commandsFile       = "commands.toml"
)

var (
	ErrNoCommands         = errors.New("no commands")
	ErrNoState            = errors.New("no state passed to control")
	ErrCmdFailed          = errors.New("could not execute command for control")
	ErrParseCmd           = errors.New("could not parse command-line")
	ErrUnknownSwitchState = errors.New("could not determine state of switch")
	ErrUnknownNumberState = errors.New("could not determine state of number")
)

// Command represents a Command to run by a button or switch.
type Command struct {
	// Name is display name for the command.
	Name string `toml:"name"`
	// Exec is the actual binary or script to run.
	Exec string `toml:"exec"`
	// Icon is a material design icon representing the command.
	Icon string `toml:"icon,omitempty"`
	// Display represents how the entity will be shown in Home Assistant. It is
	// only relevant for certain types, such as numbers and is ignored if
	// unused.
	Display string `toml:"display,omitempty"`
	// NumberType is the type of number for number controls. It should be either
	// 'int' or 'float'.
	NumberType string `toml:"type,omitempty"`
	// Min is the minimum value of the control. It is only
	// relevant for certain types, such as numbers and is ignored if unused.
	Min any `toml:"min,omitempty"`
	// Max is the maximum value of the control. It is only relevant for certain
	// types, such as numbers and is ignored if unused.
	Max any `toml:"max,omitempty"`
	// Step is the amount to change the value. It is only relevant for certain
	// types, such as numbers and is ignored if unused.
	Step any `toml:"step,omitempty"`
}

// CommandList is a CommandList of all the buttons/commands parsed from the config file.
//
//revive:disable:struct-tag
type CommandList struct {
	Buttons  []Command `toml:"button,omitempty"`
	Switches []Command `toml:"switch,omitempty"`
	Numbers  []Command `toml:"number,omitempty"`
}

// Worker represents an object with one or more buttons and switches
// definitions, which can be passed to Home Assistant to add appropriate
// entities to control the buttons/switches over MQTT.
type Worker struct {
	logger       *slog.Logger
	device       *mqtthass.Device
	buttons      []*mqtthass.ButtonEntity
	switches     []*mqtthass.SwitchEntity
	intNumbers   []*mqtthass.NumberEntity[int]
	floatNumbers []*mqtthass.NumberEntity[float64]
}

// entity is a convienience interface to avoid duplicating a lot of loops when
// configuring the controller.
type entity interface {
	MarshalSubscription() (*mqttapi.Subscription, error)
	MarshalConfig() (*mqttapi.Msg, error)
}

// Subscriptions are the MQTT subscriptions for buttons and switches, providing
// the appropriate callback mechanism to execute the associated commands.
//
//nolint:dupl
func (d *Worker) Subscriptions() []*mqttapi.Subscription {
	total := len(d.buttons) + len(d.switches) + len(d.intNumbers) + len(d.floatNumbers)
	subs := make([]*mqttapi.Subscription, 0, total)

	// Create subscriptions for buttons.
	for _, but := range d.buttons {
		subs = append(subs, d.generateSubscriptions(but))
	}
	// Create subscriptions for switches.
	for _, sw := range d.switches {
		subs = append(subs, d.generateSubscriptions(sw))
	}
	// Create subscriptions for int numbers.
	for _, inum := range d.intNumbers {
		subs = append(subs, d.generateSubscriptions(inum))
	}
	// Create subscriptions for float numbers.
	for _, fnum := range d.floatNumbers {
		subs = append(subs, d.generateSubscriptions(fnum))
	}

	return subs
}

func (d *Worker) generateSubscriptions(e entity) *mqttapi.Subscription {
	sub, err := e.MarshalSubscription()
	if err != nil {
		d.logger.Warn("Could not create entity subscription.", slog.Any("error", err))

		return nil
	}

	return sub
}

// Configs are the MQTT configurations required by Home Assistant to set up
// entities for the buttons/switches.
//
//nolint:dupl
func (d *Worker) Configs() []*mqttapi.Msg {
	total := len(d.buttons) + len(d.switches) + len(d.intNumbers) + len(d.floatNumbers)
	cfgs := make([]*mqttapi.Msg, 0, total)

	// Create button configs.
	for _, but := range d.buttons {
		cfgs = append(cfgs, d.generateConfigs(but))
	}
	// Create switch configs.
	for _, sw := range d.switches {
		cfgs = append(cfgs, d.generateConfigs(sw))
	}
	// Create int number configs.
	for _, inum := range d.intNumbers {
		cfgs = append(cfgs, d.generateConfigs(inum))
	}
	// Create float number configs.
	for _, fnum := range d.floatNumbers {
		cfgs = append(cfgs, d.generateConfigs(fnum))
	}

	return cfgs
}

func (d *Worker) generateConfigs(e entity) *mqttapi.Msg {
	msg, err := e.MarshalConfig()
	if err != nil {
		d.logger.Warn("Could not create entity config.", slog.Any("error", err))

		return nil
	}

	return msg
}

// Msgs are additional MQTT messages to be published based on any event logic
// managed by the controller. This is unused.
//
//revive:disable:unused-receiver
func (d *Worker) Msgs() chan *mqttapi.Msg {
	return nil
}

// NewCommandsWorker is used by the agent to initialize the commands
// controller, which holds the MQTT configuration for the commands defined by
// the user.
func NewCommandsWorker(ctx context.Context, device *mqtthass.Device) (*Worker, error) {
	commandsFile := filepath.Join(preferences.PathFromCtx(ctx), commandsFile)

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

	controller := &Worker{
		logger: logging.FromContext(ctx).WithGroup("custom_commands"),
		device: device,
	}
	controller.generateButtons(cmds.Buttons)
	controller.generateSwitches(cmds.Switches)
	controller.generateNumbers(cmds.Numbers)

	return controller, nil
}

// generateButtons will create MQTT entities for buttons defined by the
// controller.
func (d *Worker) generateButtons(buttonCmds []Command) {
	var id, icon, name string

	entities := make([]*mqtthass.ButtonEntity, 0, len(buttonCmds))

	for _, cmd := range buttonCmds {
		callback := func(_ *paho.Publish) {
			err := cmdWithoutState(cmd.Exec)
			if err != nil {
				d.logger.Warn("Button press failed.",
					slog.String("button", cmd.Name),
					slog.Any("error", err))
			}
		}
		name = cmd.Name
		id = strcase.ToSnake(d.device.Name + "_" + cmd.Name)

		if cmd.Icon != "" {
			icon = cmd.Icon
		} else {
			icon = "mdi:button-pointer"
		}

		entities = append(entities,
			mqtthass.NewButtonEntity().
				WithDetails(
					mqtthass.App(preferences.AppName),
					mqtthass.Name(name),
					mqtthass.ID(id),
					mqtthass.OriginInfo(preferences.MQTTOrigin()),
					mqtthass.DeviceInfo(d.device),
					mqtthass.Icon(icon),
				).WithCommand(
				mqtthass.CommandCallback(callback),
			))
	}

	d.buttons = entities
}

// generateSwitches will create MQTT entities for buttons defined by the
// controller.
func (d *Worker) generateSwitches(switchCmds []Command) {
	var id, icon, name string

	entities := make([]*mqtthass.SwitchEntity, 0, len(switchCmds))

	for _, cmd := range switchCmds {
		cmdCallBack := func(p *paho.Publish) {
			state := string(p.Payload)

			err := cmdWithState(cmd.Exec, state)
			if err != nil {
				d.logger.Warn("Switch toggle failed.",
					slog.String("switch", cmd.Name),
					slog.Any("error", err))
			}
		}
		stateCallBack := func(_ ...any) (json.RawMessage, error) {
			return switchState(cmd.Exec)
		}
		name = cmd.Name
		id = strcase.ToSnake(d.device.Name + "_" + cmd.Name)

		if cmd.Icon != "" {
			icon = cmd.Icon
		} else {
			icon = "mdi:toggle-switch"
		}

		entities = append(entities,
			mqtthass.NewSwitchEntity().
				WithDetails(
					mqtthass.App(preferences.AppName),
					mqtthass.Name(name),
					mqtthass.ID(id),
					mqtthass.OriginInfo(preferences.MQTTOrigin()),
					mqtthass.DeviceInfo(d.device),
					mqtthass.Icon(icon),
				).
				WithCommand(
					mqtthass.CommandCallback(cmdCallBack),
				).
				WithState(
					mqtthass.StateCallback(stateCallBack),
				).
				OptimisticMode())
	}

	d.switches = entities
}

// generateNumbers will create MQTT entities for numbers (both ints and floats) defined by the
// controller.
//
//nolint:gocognit,funlen
func (d *Worker) generateNumbers(numberCommands []Command) {
	var (
		id, icon, name string
		ints           []*mqtthass.NumberEntity[int]
		floats         []*mqtthass.NumberEntity[float64]
	)

	for _, cmd := range numberCommands {
		cmdCallBack := func(p *paho.Publish) {
			state := string(p.Payload)

			err := cmdWithState(cmd.Exec, state)
			if err != nil {
				d.logger.Warn("Set number failed.",
					slog.String("number", cmd.Name),
					slog.Any("error", err))
			}
		}
		stateCallBack := func(_ ...any) (json.RawMessage, error) {
			return numberState(cmd.Exec)
		}
		name = cmd.Name
		id = strcase.ToSnake(d.device.Name + "_" + cmd.Name)

		if cmd.Icon != "" {
			// Set the icon to the user-specified icon
			icon = cmd.Icon
		} else {
			// Choose an appropriate icon based on the display value.
			switch cmd.Display {
			case "box":
				icon = "mdi:counter"
			case "slider":
				icon = "mdi:tune"
			default:
				icon = "mdi:knob"
			}
		}

		// Set the display type based on any configuration specified. Else,
		// default to "auto".
		displayType := mqtthass.NumberAuto

		switch cmd.Display {
		case "box":
			displayType = mqtthass.NumberBox
		case "slider":
			displayType = mqtthass.NumberSlider
		}

		// Add an entity based on the number type.
		valueType := cmd.NumberType

		switch valueType {
		case "float":
			min := convValue[float64](cmd.Min) //nolint:predeclared

			max := convValue[float64](cmd.Max) //nolint:predeclared
			if max == 0 {
				max = 100
			}

			step := convValue[float64](cmd.Step)
			if step == 0 {
				step = 1
			}

			floats = append(floats,
				mqtthass.NewNumberEntity[float64]().
					WithDetails(
						mqtthass.App(preferences.AppName),
						mqtthass.Name(name),
						mqtthass.ID(id),
						mqtthass.OriginInfo(preferences.MQTTOrigin()),
						mqtthass.DeviceInfo(d.device),
						mqtthass.Icon(icon),
					).
					WithCommand(
						mqtthass.CommandCallback(cmdCallBack),
					).
					WithState(
						mqtthass.StateCallback(stateCallBack),
						mqtthass.ValueTemplate(stateValueTemplate),
					).
					WithMode(displayType).
					WithStep(step).
					WithMin(min).
					WithMax(max).
					OptimisticMode())

		default:
			min := convValue[int](cmd.Min) //nolint:predeclared

			max := convValue[int](cmd.Max) //nolint:predeclared
			if max == 0 {
				max = 100
			}

			step := convValue[int](cmd.Step)
			if step == 0 {
				step = 1
			}

			ints = append(ints,
				mqtthass.NewNumberEntity[int]().
					WithDetails(
						mqtthass.App(preferences.AppName),
						mqtthass.Name(name),
						mqtthass.ID(id),
						mqtthass.OriginInfo(preferences.MQTTOrigin()),
						mqtthass.DeviceInfo(d.device),
						mqtthass.Icon(icon),
					).
					WithCommand(
						mqtthass.CommandCallback(cmdCallBack),
					).
					WithState(
						mqtthass.StateCallback(stateCallBack),
						mqtthass.ValueTemplate(stateValueTemplate),
					).
					WithMode(displayType).
					WithStep(step).
					WithMin(min).
					WithMax(max).
					OptimisticMode())
		}
	}

	d.floatNumbers = floats
	d.intNumbers = ints
}

// cmdWithoutState runs the executable associated with a control with no state
// passed to the command. This is used for controls which do not have a state,
// like buttons.
func cmdWithoutState(command string) error {
	cmdElems := strings.Split(command, " ")

	if len(cmdElems) == 0 {
		return ErrParseCmd
	}

	_, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCmdFailed, err)
	}

	return nil
}

// cmdWithState runs the executable associated with a control, passing a state
// value. This is used by controls with a controllable state in Home Assistant.
func cmdWithState(command, state string) error {
	if state == "" {
		return ErrNoState
	}

	cmdElems := strings.Split(command, " ")
	cmdElems = append(cmdElems, state)

	_, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCmdFailed, err)
	}

	return nil
}

// switchState will execute the command associated with the switch control,
// which should output the current state of the switch.
func switchState(command string) (json.RawMessage, error) {
	cmdElems := strings.Split(command, " ")
	if len(cmdElems) == 0 {
		return nil, ErrUnknownSwitchState
	}

	output, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return nil, fmt.Errorf("could get switch state: %w", err)
	}

	switch {
	case bytes.Contains(output, []byte(switchOnState)):
		return json.RawMessage(switchOnState), nil
	case bytes.Contains(output, []byte(switchOffState)):
		return json.RawMessage(switchOffState), nil
	}

	return nil, ErrUnknownSwitchState
}

func numberState(command string) (json.RawMessage, error) {
	cmdElems := strings.Split(command, " ")
	if len(cmdElems) == 0 {
		return nil, ErrUnknownNumberState
	}

	output, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnknownNumberState, err)
	}

	number := string(bytes.TrimSpace(output))

	_, err1 := strconv.ParseInt(number, 10, 64)
	_, err2 := strconv.ParseFloat(number, 64)

	if err1 != nil && err2 != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnknownNumberState, errors.Join(err1, err2))
	}

	return json.RawMessage(`{ "value": ` + number + ` }`), nil
}

// convValue provides a generic way to either convert to an int/float or just
// return the default value of that type.
func convValue[T constraints.Float | constraints.Integer](orig any) T {
	value, ok := orig.(T)
	if !ok {
		return T(0)
	}

	return value
}
