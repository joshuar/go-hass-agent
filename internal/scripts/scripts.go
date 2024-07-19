// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package scripts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

type Script struct {
	path     string
	schedule string
}

func (s *Script) Schedule() string {
	return s.schedule
}

func (s *Script) Execute() ([]sensor.Details, error) {
	output, err := s.parse()
	if err != nil {
		return nil, fmt.Errorf("error running script: %w", err)
	}

	sensors := make([]sensor.Details, 0, len(output.Sensors))

	for _, s := range output.Sensors {
		sensors = append(sensors, sensor.Details(&s))
	}

	return sensors, nil
}

//nolint:exhaustruct
func (s *Script) parse() (*scriptOutput, error) {
	cmdElems := strings.Split(s.path, " ")

	out, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return nil, fmt.Errorf("could not execute script: %w", err)
	}

	output := &scriptOutput{}

	err = output.Unmarshal(out)
	if err != nil {
		return nil, fmt.Errorf("could not parse script output: %w", err)
	}

	return output, nil
}

// NewScript returns a new script object that can scheduled with the job
// scheduler by the agent.
func NewScript(path string) (*Script, error) {
	script := &Script{
		path:     path,
		schedule: "",
	}

	scriptOutput, err := script.parse()
	if err != nil {
		return nil, fmt.Errorf("cannot add script %s: %w", path, err)
	}

	script.schedule = scriptOutput.Schedule

	return script, nil
}

// scriptOutput represents the output from a script. The output must be
// formatted as either valid JSON or YAML. This output is used to define a
// sensor in Home Assistant.
//
//nolint:tagalign
type scriptOutput struct {
	Schedule string         `json:"schedule" yaml:"schedule"`
	Sensors  []ScriptSensor `json:"sensors" yaml:"sensors"`
}

// Unmarshal will attempt to take the raw output from a script execution and
// format it as either JSON or YAML. If successful, this format can then be used
// as a sensor.
func (o *scriptOutput) Unmarshal(scriptOutput []byte) error {
	jsonErr := json.Unmarshal(scriptOutput, &o)
	if jsonErr == nil {
		return nil
	}

	yamlErr := yaml.Unmarshal(scriptOutput, &o)
	if yamlErr == nil {
		return nil
	}

	tomlErr := toml.Unmarshal(scriptOutput, &o)
	if tomlErr == nil {
		return nil
	}

	return errors.Join(jsonErr, yamlErr, tomlErr)
}

//nolint:tagalign
type ScriptSensor struct {
	SensorState       any    `json:"sensor_state" yaml:"sensor_state" toml:"sensor_state"`
	SensorAttributes  any    `json:"sensor_attributes,omitempty" yaml:"sensor_attributes,omitempty" toml:"sensor_attributes,omitempty"`
	SensorName        string `json:"sensor_name" yaml:"sensor_name" toml:"sensor_name"`
	SensorIcon        string `json:"sensor_icon,omitempty" yaml:"sensor_icon,omitempty" toml:"sensor_icon,omitempty"`
	SensorDeviceClass string `json:"sensor_device_class,omitempty" yaml:"sensor_device_class,omitempty" toml:"sensor_device_class,omitempty"`
	SensorStateClass  string `json:"sensor_state_class,omitempty" yaml:"sensor_state_class,omitempty" toml:"sensor_state_class,omitempty"`
	SensorStateType   string `json:"sensor_type,omitempty" yaml:"sensor_type,omitempty" toml:"sensor_type,omitempty"`
	SensorUnits       string `json:"sensor_units,omitempty" yaml:"sensor_units,omitempty" toml:"sensor_units,omitempty"`
}

func (s *ScriptSensor) Name() string {
	return s.SensorName
}

func (s *ScriptSensor) ID() string {
	return strcase.ToSnake(s.SensorName)
}

func (s *ScriptSensor) Icon() string {
	if s.SensorIcon == "" {
		return "mdi:script"
	}

	return s.SensorIcon
}

func (s *ScriptSensor) SensorType() types.SensorClass {
	switch s.SensorStateType {
	case "binary":
		return types.BinarySensor
	default:
		return types.Sensor
	}
}

func (s *ScriptSensor) DeviceClass() types.DeviceClass {
	for d := types.DeviceClassApparentPower; d <= types.DeviceClassWindSpeed; d++ {
		if s.SensorDeviceClass == d.String() {
			return d
		}
	}

	return 0
}

func (s *ScriptSensor) StateClass() types.StateClass {
	switch s.SensorStateClass {
	case "measurement":
		return types.StateClassMeasurement
	case "total":
		return types.StateClassTotal
	case "total_increasing":
		return types.StateClassTotalIncreasing
	default:
		return 0
	}
}

func (s *ScriptSensor) State() any {
	return s.SensorState
}

func (s *ScriptSensor) Units() string {
	return s.SensorUnits
}

func (s *ScriptSensor) Category() string {
	return ""
}

func (s *ScriptSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	if s.SensorAttributes != nil {
		attributes["extra_attributes"] = s.SensorAttributes
	}

	return attributes
}
