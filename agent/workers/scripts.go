// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package workers

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
	"sync"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml/v2"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"
	"gopkg.in/yaml.v3"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/id"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/scheduler"
)

var (
	// ErrExecuteScript represents an error that occurred when running a script.
	ErrExecuteScript = errors.New("could not execute script")
	// ErrParseCmd represents an error that occurred trying to parse a script command-line.
	ErrParseCmd = errors.New("could not parse script command")
)

// Verify Script satisfies the quartz job interface.
var _ quartz.Job = (*Script)(nil)

// Script represents a script.
type Script struct {
	path     string
	schedule string
	outCh    chan models.Entity
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
			slogctx.FromCtx(ctx).Warn("Could not execute script.",
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
		s.outCh <- scriptToEntity(ctx, sensor)
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
		sensors = append(sensors, scriptToEntity(ctx, s))
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

// ErrNewSensor is returned when a problem occurred creating a sensor entity.
var ErrNewSensor = errors.New("could not create sensor entity")

// ScriptSensor represents a sensor generated from script output.
type ScriptSensor struct {
	SensorState       any            `json:"sensor_state" yaml:"sensor_state" toml:"sensor_state"`
	SensorAttributes  map[string]any `json:"sensor_attributes,omitempty" yaml:"sensor_attributes,omitempty" toml:"sensor_attributes,omitempty"`
	SensorName        string         `json:"sensor_name" yaml:"sensor_name" toml:"sensor_name"`
	SensorIcon        string         `json:"sensor_icon,omitempty" yaml:"sensor_icon,omitempty" toml:"sensor_icon,omitempty"`
	SensorDeviceClass string         `json:"sensor_device_class,omitempty" yaml:"sensor_device_class,omitempty" toml:"sensor_device_class,omitempty"`
	SensorStateClass  string         `json:"sensor_state_class,omitempty" yaml:"sensor_state_class,omitempty" toml:"sensor_state_class,omitempty"`
	SensorStateType   string         `json:"sensor_type,omitempty" yaml:"sensor_type,omitempty" toml:"sensor_type,omitempty"`
	SensorUnits       string         `json:"sensor_units,omitempty" yaml:"sensor_units,omitempty" toml:"sensor_units,omitempty"`
}

func scriptToEntity(ctx context.Context, script ScriptSensor) models.Entity {
	var typeOption sensor.Option

	switch script.SensorStateType {
	case "binary":
		typeOption = sensor.AsTypeBinarySensor()
	default:
		typeOption = sensor.AsTypeSensor()
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(script.SensorName),
		sensor.WithID(strcase.ToSnake(script.SensorName)),
		sensor.WithUnits(script.SensorUnits),
		sensor.WithDeviceClass(script.DeviceClass()),
		sensor.WithStateClass(script.StateClass()),
		sensor.WithIcon(script.Icon()),
		sensor.WithAttributes(script.Attributes()),
		sensor.WithState(script.SensorState),
		typeOption,
	)
}

// Icon is an material design icon to represent the script state.
func (s *ScriptSensor) Icon() string {
	if s.SensorIcon == "" {
		return "mdi:script"
	}

	return s.SensorIcon
}

// DeviceClass is a sensor device class for the script state.
func (s *ScriptSensor) DeviceClass() class.SensorDeviceClass {
	for d := class.SensorClassMin + 1; d <= class.BinaryClassMax; d++ {
		if s.SensorDeviceClass == d.String() {
			return d
		}
	}

	return 0
}

// StateClass is a sensor state class for the script state.
func (s *ScriptSensor) StateClass() class.SensorStateClass {
	switch s.SensorStateClass {
	case "measurement":
		return class.StateMeasurement
	case "total":
		return class.StateTotal
	case "total_increasing":
		return class.StateTotalIncreasing
	default:
		return class.StateClassMin
	}
}

// Attributes are any additional custom attributes for the script state.
func (s *ScriptSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	if s.SensorAttributes != nil {
		attributes = s.SensorAttributes
	}

	return attributes
}

var (
	ErrUnknownScript    = errors.New("unknown or nonexistent script")
	ErrAlreadyStarted   = errors.New("script already started")
	ErrAlreadyStopped   = errors.New("script already stopped")
	ErrSchedulingFailed = errors.New("failed to schedule script")
	ErrParseSchedule    = errors.New("could not parse script schedule")
)

const (
	scriptWorkerID   = "scripts"
	scriptWorkerDesc = "Custom script-based sensors"
)

// ScriptWorker is a worker for custom scripts.
type ScriptWorker struct {
	*models.WorkerMetadata

	scripts []*Script
	outCh   chan models.Entity
	prefs   *CommonWorkerPrefs
}

// NewScriptsWorker creates a new worker for custom scripts.
func NewScriptsWorker(ctx context.Context) (*ScriptWorker, error) {
	scriptPath := filepath.Join(config.GetPath(), "scripts")

	worker := &ScriptWorker{
		WorkerMetadata: models.SetWorkerMetadata(scriptWorkerID, scriptWorkerDesc),
	}

	defaultPrefs := &CommonWorkerPrefs{}

	scripts, err := worker.findScripts(ctx, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("could not find scripts: %w", err)
	}
	worker.scripts = scripts

	worker.prefs, err = LoadWorkerPreferences("scripts", defaultPrefs)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	return worker, nil
}

// IsDisabled returns a boolean indicating whether the scripts worker is disabled.
func (c *ScriptWorker) IsDisabled() bool {
	return c.prefs.IsDisabled()
}

// States will execute all running scripts and returns their sensor entities.
func (c *ScriptWorker) States(ctx context.Context) []models.Entity {
	var allSensors []models.Entity

	for _, script := range c.scripts {
		scriptSensors, err := script.Run(ctx)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not retrieve script sensors",
				slog.String("script", script.Description()),
				slog.Any("error", err),
			)

			continue
		}

		allSensors = append(allSensors, scriptSensors...)
	}

	return allSensors
}

func (c *ScriptWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	c.outCh = make(chan models.Entity)

	scriptOutputs := make([]<-chan models.Entity, 0, len(c.scripts))

	for _, script := range c.scripts {
		// Parse the script cron schedule as a scheduler trigger.
		trigger, err := parseSchedule(script.Schedule())
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not schedule script.",
				slog.String("script", script.Description()),
				slog.Any("error", err))

			continue
		}
		// Schedule the script.
		err = scheduler.Manager.ScheduleJob(id.ScriptJob, script, trigger)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not schedule script.",
				slog.String("script", script.Description()),
				slog.Any("error", err))

			continue
		}
		// Append to list of managed scripts.
		scriptOutputs = append(scriptOutputs, script.Start(ctx))
	}

	return mergeCh(ctx, scriptOutputs...), nil
}

func (c *ScriptWorker) Stop() error {
	close(c.outCh)

	return nil
}

// findScripts locates scripts and returns a slice of scripts that the agent can
// run.
func (c *ScriptWorker) findScripts(ctx context.Context, path string) ([]*Script, error) {
	var sensorScripts []*Script

	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not search for scripts: %w", err)
	}

	for _, scriptFile := range files {
		if isExecutable(scriptFile) {
			script, err := NewScript(scriptFile)
			if err != nil {
				slogctx.FromCtx(ctx).Warn("Script error.",
					slog.Any("error", err),
				)

				continue
			}

			sensorScripts = append(sensorScripts, script)
		}
	}

	return sensorScripts, nil
}

// isExecutable is helper to determine if a (script) file is executable.
func isExecutable(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.Mode().Perm()&0o111 != 0
}

// mergeCh merges a list of channels of any type into a single channel of that
// type (channel fan-in).
func mergeCh[T any](ctx context.Context, inCh ...<-chan T) chan T {
	var wg sync.WaitGroup

	outCh := make(chan T)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(ch <-chan T) { //nolint:varnamelen
		defer wg.Done()

		if ch == nil {
			return
		}

		for n := range ch {
			select {
			case outCh <- n:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(inCh))

	for _, c := range inCh {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

// parseSchedule parses a cron schedule string and returns the equivalent quartz
// Trigger.
//
// Cron schedule parsing code adapted from
// https://github.com/robfig/cron/blob/master/parser.go
func parseSchedule(sched string) (quartz.Trigger, error) {
	var (
		trigger quartz.Trigger
		err     error
	)

	// Attempt to parse as a standard cron schedule string.
	trigger, err = quartz.NewCronTrigger(sched)
	if err == nil {
		return trigger, nil
	}

	// Attempt to parse as one of the year/month/week/day/hour strings.
	switch sched {
	case "@yearly", "@annually":
		trigger, err = quartz.NewCronTrigger("0 0 0 1 1 * *")
	case "@monthly":
		trigger, err = quartz.NewCronTrigger("0 0 0 1 * *")
	case "@weekly":
		trigger, err = quartz.NewCronTrigger("0 0 0 * * 1")
	case "@daily", "@midnight":
		trigger, err = quartz.NewCronTrigger("0 0 0 * * *")
	case "@hourly":
		trigger, err = quartz.NewCronTrigger("0 0 * * * *")
	}
	// If successfully parsed, return the trigger.
	if err == nil {
		return trigger, nil
	}

	// Else, attempt to parse as an "@every ..." string.
	const every = "@every "
	if strings.HasPrefix(sched, every) {
		duration, err := time.ParseDuration(sched[len(every):])
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParseSchedule, err)
		}

		return quartz.NewSimpleTrigger(duration), nil
	}

	return nil, fmt.Errorf("%w: unknown schedule format %s", ErrParseSchedule, sched)
}
