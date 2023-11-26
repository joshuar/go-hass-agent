// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package scripts

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type script struct {
	path     string
	schedule string
	Output   chan tracker.Sensor
}

func (s *script) execute() (*scriptOutput, error) {
	cmd := exec.Command(s.path)
	o, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	output := &scriptOutput{}
	err = output.Unmarshal(o)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Run is the function that is called when a script is run by the scheduler on
// its specified schedule. It is implemented to satisfy the cron package
// interface, so the script can be treated as a cron job. Run will execute the
// script, collect the output and send it through a channel as a sensor object.
func (s *script) Run() {
	output, err := s.execute()
	if err != nil {
		log.Warn().Err(err).Str("script", s.path).
			Msg("Could not run script.")
		return
	}

	for _, sensor := range output.Sensors {
		s.Output <- sensor
	}
}

// Schedule retrieves the cron schedule that the script should be run on.73
func (s *script) Schedule() string {
	return s.schedule
}

// Path returns the path to the script on disk
func (s *script) Path() string {
	return s.path
}

// NewScript returns a new script object that can scheduled with the joib
// scheduler by the agent.
func NewScript(p string) *script {
	s := &script{
		path:   p,
		Output: make(chan tracker.Sensor),
	}
	o, err := s.execute()
	if err != nil {
		log.Warn().Err(err).Str("script", p).
			Msg("Cannot run script")
		return nil
	}
	s.schedule = o.Schedule
	return s
}

// scriptOutput represents the output from a script. The output must be
// formatted as either valid JSON or YAML. This output is used to define a
// sensor in Home Assistant.
type scriptOutput struct {
	Schedule string          `json:"schedule" yaml:"schedule"`
	Sensors  []*scriptSensor `json:"sensors" yaml:"sensors"`
}

// Unmarshal will attempt to take the raw output from a script execution and
// format it as either JSON or YAML. If successful, this format can then be used
// as a sensor.
func (o *scriptOutput) Unmarshal(b []byte) error {
	var err error
	err = json.Unmarshal(b, &o)
	if err == nil {
		return nil
	}
	err = yaml.Unmarshal(b, &o)
	if err == nil {
		return nil
	}
	return errors.New("could not unmarshal script output")
}

type scriptSensor struct {
	SensorName        string      `json:"sensor_name" yaml:"sensor_name"`
	SensorIcon        string      `json:"sensor_icon" yaml:"sensor_icon"`
	SensorDeviceClass string      `json:"sensor_device_class,omitempty" yaml:"sensor_device_class,omitempty"`
	SensorStateClass  string      `json:"sensor_state_class,omitempty" yaml:"sensor_state_class,omitempty"`
	SensorStateType   string      `json:"sensor_type,omitempty" yaml:"sensor_type,omitempty"`
	SensorState       interface{} `json:"sensor_state" yaml:"sensor_state"`
	SensorUnits       string      `json:"sensor_units,omitempty" yaml:"sensor_units,omitempty"`
	SensorAttributes  interface{} `json:"sensor_attributes,omitempty" yaml:"sensor_attributes,omitempty"`
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

func (s *scriptSensor) SensorType() sensor.SensorType {
	switch s.SensorStateType {
	case "binary":
		return sensor.TypeBinary
	default:
		return sensor.TypeSensor
	}
}

func (s *scriptSensor) DeviceClass() sensor.SensorDeviceClass {
	for d := sensor.Apparent_power; d <= sensor.Wind_speed; d++ {
		if s.SensorDeviceClass == d.String() {
			return d
		}
	}
	return 0
}

func (s *scriptSensor) StateClass() sensor.SensorStateClass {
	switch s.SensorStateClass {
	case "measurement":
		return sensor.StateMeasurement
	case "total":
		return sensor.StateTotal
	case "total_increasing":
		return sensor.StateTotalIncreasing
	default:
		return 0
	}
}

func (s *scriptSensor) State() interface{} {
	return s.SensorState
}

func (s *scriptSensor) Units() string {
	return s.SensorUnits
}

func (s *scriptSensor) Category() string {
	return ""
}

func (s *scriptSensor) Attributes() interface{} {
	return s.SensorAttributes
}

func FindScripts(path string) ([]*script, error) {
	var scripts []*script
	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, err
	}
	for _, s := range files {
		if isExecutable(s) {
			scripts = append(scripts, NewScript(s))
		}
	}
	return scripts, nil
}

func isExecutable(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return fi.Mode().Perm()&0111 != 0
}
