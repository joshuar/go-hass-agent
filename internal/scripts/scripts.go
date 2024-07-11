// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package scripts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type Script struct {
	Output   chan sensor.Details
	logger   *slog.Logger
	path     string
	schedule string
}

//nolint:exhaustruct
func (s *Script) execute() (*scriptOutput, error) {
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

// Run is the function that is called when a script is run by the scheduler on
// its specified schedule. It is implemented to satisfy the cron package
// interface, so the script can be treated as a cron job. Run will execute the
// script, collect the output and send it through a channel as a sensor object.
func (s *Script) Run() {
	output, err := s.execute()
	if err != nil {
		s.logger.Warn("Could not execute script.", "script", s.path, "error", err.Error())

		return
	}

	for _, o := range output.Sensors {
		s.Output <- o
	}
}

// Schedule retrieves the cron schedule that the script should be run on.
func (s *Script) Schedule() string {
	return s.schedule
}

// Path returns the path to the script on disk.
func (s *Script) Path() string {
	return s.path
}

// NewScript returns a new script object that can scheduled with the job
// scheduler by the agent.
func NewScript(ctx context.Context, path string) (*Script, error) {
	script := &Script{
		path:     path,
		schedule: "",
		Output:   make(chan sensor.Details),
		logger:   logging.FromContext(ctx).With(slog.String("component", "scripts")),
	}

	out, err := script.execute()
	if err != nil {
		return nil, fmt.Errorf("cannot add script %s: %w", path, err)
	}

	script.schedule = out.Schedule

	return script, nil
}

// scriptOutput represents the output from a script. The output must be
// formatted as either valid JSON or YAML. This output is used to define a
// sensor in Home Assistant.
//
//nolint:tagalign
type scriptOutput struct {
	Schedule string          `json:"schedule" yaml:"schedule"`
	Sensors  []*scriptSensor `json:"sensors" yaml:"sensors"`
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
type scriptSensor struct {
	SensorState       any    `json:"sensor_state" yaml:"sensor_state" toml:"sensor_state"`
	SensorAttributes  any    `json:"sensor_attributes,omitempty" yaml:"sensor_attributes,omitempty" toml:"sensor_attributes,omitempty"`
	SensorName        string `json:"sensor_name" yaml:"sensor_name" toml:"sensor_name"`
	SensorIcon        string `json:"sensor_icon" yaml:"sensor_icon" toml:"sensor_icon"`
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

	attributes["extra_attributes"] = s.SensorAttributes

	return attributes
}

// FindScripts locates scripts and returns a slice of scripts that the agent can
// run.
func FindScripts(ctx context.Context, path string) ([]*Script, error) {
	var scripts []*Script

	var errs error

	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not search for scripts: %w", err)
	}

	for _, scriptFile := range files {
		if isExecutable(scriptFile) {
			script, err := NewScript(ctx, scriptFile)
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
