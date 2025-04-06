// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package scripts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/reugn/go-quartz/quartz"
	"gopkg.in/yaml.v3"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/models"
)

var (
	// ErrExecuteScript represents an error that occurred when running a script.
	ErrExecuteScript = errors.New("could not execute script")
	// ErrParseCmd represents an error that occurred trying to parse a script command-line.
	ErrParseCmd = errors.New("could not parse script command")
)

// Verify Script satisfies the quartz job interface.
var _ quartz.Job = (*Script)(nil)

type Script struct {
	path     string
	schedule string
	outCh    chan models.Entity
}

// Start is used to create the output channel for the script. It will also
// Execute the script once.
func (s *Script) Start(ctx context.Context) <-chan models.Entity {
	// Create the channel for script output.
	s.outCh = make(chan models.Entity)
	// Clean-up on agent close.
	go func() {
		defer close(s.outCh)
		<-ctx.Done()
	}()
	// Send initial update.
	go func() {
		if err := s.Execute(ctx); err != nil {
			logging.FromContext(ctx).Warn("Could not execute script.",
				slog.Any("error", err))
		}
	}()

	return s.outCh
}

// Schedule returns the script's cron schedule string.
func (s *Script) Schedule() string {
	return s.schedule
}

// Execute will run the script and send any sensor entities it outputs to the
// script's output channel. If there was an error running the script, a non-nil
// error is returned.
func (s *Script) Execute(ctx context.Context) error {
	output, err := s.parse()
	if err != nil {
		return errors.Join(ErrExecuteScript, err)
	}

	for _, sensor := range output.Sensors {
		entity, err := scriptToEntity(ctx, sensor)
		if err != nil {
			return errors.Join(ErrExecuteScript, err)
		}

		s.outCh <- *entity
	}

	return nil
}

// Run will run the script and return a slice of sensor entities. If there was
// an error running the script, a non-nil error is returned.
func (s *Script) Run(ctx context.Context) ([]models.Entity, error) {
	output, err := s.parse()
	if err != nil {
		return nil, fmt.Errorf("error running script: %w", err)
	}

	sensors := make([]models.Entity, 0, len(output.Sensors))

	for _, s := range output.Sensors {
		entity, err := scriptToEntity(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("could not create script entity: %w", err)
		}

		sensors = append(sensors, *entity)
	}

	return sensors, nil
}

// Description returns a formatted string showing the script path and schedule.
func (s *Script) Description() string {
	return fmt.Sprintf("Run %s on schedule %s", s.path, s.schedule)
}

// parse extracts the script schedule and sensor output from the script output.
// It will return a non-nil error if there was a problem parsing the script output.
func (s *Script) parse() (*scriptOutput, error) {
	cmdElems := strings.Split(s.path, " ")

	if len(cmdElems) == 0 {
		return nil, ErrParseCmd
	}

	out, err := exec.Command(cmdElems[0], cmdElems[1:]...).Output() // #nosec: G204
	if err != nil {
		return nil, fmt.Errorf("could not execute script: %w", err)
	}

	output := &scriptOutput{}

	if err := output.Unmarshal(out); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
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
type scriptOutput struct {
	Schedule string         `json:"schedule" yaml:"schedule"`
	Sensors  []ScriptSensor `json:"sensors" yaml:"sensors"`
}

// Unmarshal will attempt to take the raw output from a script execution and
// format it as either JSON or YAML. If successful, this format can then be used
// as a models.
//
//revive:disable:indent-error-flow // errors need to be gathered.
func (o *scriptOutput) Unmarshal(scriptOutput []byte) error {
	var parseError error

	if err := json.Unmarshal(scriptOutput, o); err == nil {
		return nil
	} else {
		parseError = errors.Join(parseError, fmt.Errorf("not valid JSON: %w", err))
	}

	if err := yaml.Unmarshal(scriptOutput, o); err == nil {
		return nil
	} else {
		parseError = errors.Join(parseError, fmt.Errorf("not valid YAML: %w", err))
	}

	if err := toml.Unmarshal(scriptOutput, o); err == nil {
		return nil
	} else {
		parseError = errors.Join(parseError, fmt.Errorf("not valid TOML: %w", err))
	}

	return parseError
}
