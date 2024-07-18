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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

type Script struct {
	Path     string
	Schedule string
}

//nolint:exhaustruct
func (s *Script) Execute() (*ScriptOutput, error) {
	cmdElems := strings.Split(s.Path, " ")

	out, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output()
	if err != nil {
		return nil, fmt.Errorf("could not execute script: %w", err)
	}

	output := &ScriptOutput{}

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
		Path:     path,
		Schedule: "",
	}

	out, err := script.Execute()
	if err != nil {
		return nil, fmt.Errorf("cannot add script %s: %w", path, err)
	}

	script.Schedule = out.Schedule

	return script, nil
}

// ScriptOutput represents the output from a script. The output must be
// formatted as either valid JSON or YAML. This output is used to define a
// sensor in Home Assistant.
//
//nolint:tagalign
type ScriptOutput struct {
	Schedule string          `json:"schedule" yaml:"schedule"`
	Sensors  []*scriptSensor `json:"sensors" yaml:"sensors"`
}

// Unmarshal will attempt to take the raw output from a script execution and
// format it as either JSON or YAML. If successful, this format can then be used
// as a sensor.
func (o *ScriptOutput) Unmarshal(scriptOutput []byte) error {
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
type scriptSensor struct {
	SensorState       any    `json:"sensor_state" yaml:"sensor_state" toml:"sensor_state"`
	SensorAttributes  any    `json:"sensor_attributes,omitempty" yaml:"sensor_attributes,omitempty" toml:"sensor_attributes,omitempty"`
	SensorName        string `json:"sensor_name" yaml:"sensor_name" toml:"sensor_name"`
	SensorIcon        string `json:"sensor_icon,omitempty" yaml:"sensor_icon,omitempty" toml:"sensor_icon,omitempty"`
	SensorDeviceClass string `json:"sensor_device_class,omitempty" yaml:"sensor_device_class,omitempty" toml:"sensor_device_class,omitempty"`
	SensorStateClass  string `json:"sensor_state_class,omitempty" yaml:"sensor_state_class,omitempty" toml:"sensor_state_class,omitempty"`
	SensorStateType   string `json:"sensor_type,omitempty" yaml:"sensor_type,omitempty" toml:"sensor_type,omitempty"`
	SensorUnits       string `json:"sensor_units,omitempty" yaml:"sensor_units,omitempty" toml:"sensor_units,omitempty"`
}

func (s *scriptSensor) Name() string {
	return s.SensorName
}

func (s *scriptSensor) ID() string {
	return strcase.ToSnake(s.SensorName)
}

func (s *scriptSensor) Icon() string {
	if s.SensorIcon == "" {
		return "mdi:script"
	}

	return s.SensorIcon
}

func (s *scriptSensor) SensorType() types.SensorClass {
	switch s.SensorStateType {
	case "binary":
		return types.BinarySensor
	default:
		return types.Sensor
	}
}

func (s *scriptSensor) DeviceClass() types.DeviceClass {
	for d := types.DeviceClassApparentPower; d <= types.DeviceClassWindSpeed; d++ {
		if s.SensorDeviceClass == d.String() {
			return d
		}
	}

	return 0
}

func (s *scriptSensor) StateClass() types.StateClass {
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

func (s *scriptSensor) State() any {
	return s.SensorState
}

func (s *scriptSensor) Units() string {
	return s.SensorUnits
}

func (s *scriptSensor) Category() string {
	return ""
}

func (s *scriptSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	if s.SensorAttributes != nil {
		attributes["extra_attributes"] = s.SensorAttributes
	}

	return attributes
}

// FindScripts locates scripts and returns a slice of scripts that the agent can
// run.
func FindScripts(path string) ([]*Script, error) {
	var scripts []*Script

	var errs error

	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not search for scripts: %w", err)
	}

	for _, scriptFile := range files {
		if isExecutable(scriptFile) {
			script, err := NewScript(scriptFile)
			if err != nil {
				errs = errors.Join(errs, err)

				continue
			}

			scripts = append(scripts, script)
		}
	}

	return scripts, nil
}

func isExecutable(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.Mode().Perm()&0o111 != 0
}
