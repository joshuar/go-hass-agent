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

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

var ErrParseCmd = errors.New("could not parse script command")

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

func (s *Script) parse() (*scriptOutput, error) {
	cmdElems := strings.Split(s.path, " ")

	if len(cmdElems) == 0 {
		return nil, ErrParseCmd
	}

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
