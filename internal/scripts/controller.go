// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package scripts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/robfig/cron/v3"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

var (
	ErrUnknownScript    = errors.New("unknown or nonexistent script")
	ErrAlreadyStarted   = errors.New("script already started")
	ErrAlreadyStopped   = errors.New("script already stopped")
	ErrSchedulingFailed = errors.New("failed to schedule script")
)

type job struct {
	logAttrs slog.Attr
	Script
	ID cron.EntryID
}

type Controller struct {
	scheduler *cron.Cron
	logger    *slog.Logger
	jobs      []job
}

func (c *Controller) ID() string {
	return "scripts"
}

func (c *Controller) States(_ context.Context) []sensor.Details {
	var sensors []sensor.Details

	for _, worker := range c.ActiveWorkers() {
		found := slices.IndexFunc(c.jobs, func(j job) bool { return j.path == worker })

		jobSensors, err := c.jobs[found].Execute()
		if err != nil {
			c.logger.Warn("Could not retrieve script sensors",
				slog.String("script", c.jobs[found].path),
				slog.Any("error", err),
			)
		} else {
			sensors = append(sensors, jobSensors...)
		}
	}

	return sensors
}

func (c *Controller) ActiveWorkers() []string {
	var activeScripts []string

	for _, job := range c.jobs {
		if job.ID != 0 {
			activeScripts = append(activeScripts, job.path)
		}
	}

	return activeScripts
}

func (c *Controller) InactiveWorkers() []string {
	var inactiveScripts []string

	for _, job := range c.jobs {
		if job.ID == 0 {
			inactiveScripts = append(inactiveScripts, job.path)
		}
	}

	return inactiveScripts
}

func (c *Controller) Start(_ context.Context, name string) (<-chan sensor.Details, error) {
	found := slices.IndexFunc(c.jobs, func(j job) bool { return j.path == name })

	// If the script was not found, return an error.
	if found == -1 {
		return nil, ErrUnknownScript
	}
	// If the script is already started, return an error.
	if c.jobs[found].ID > 0 {
		return nil, ErrAlreadyStarted
	}

	sensorCh := make(chan sensor.Details)

	// Schedule the script.
	id, err := c.scheduler.AddFunc(c.jobs[found].Schedule(), func() {
		sensors, err := c.jobs[found].Execute()
		if err != nil {
			c.logger.Warn("Could not execute script.",
				c.jobs[found].logAttrs,
				slog.Any("error", err))

			return
		}

		for _, o := range sensors {
			sensorCh <- o
		}
	})
	if err != nil {
		close(sensorCh)

		return nil, ErrSchedulingFailed
	}

	// Update the job id.
	c.jobs[found].ID = id

	// Return the new sensor channel for the script.
	return sensorCh, nil
}

func (c *Controller) Stop(name string) error {
	found := slices.IndexFunc(c.jobs, func(j job) bool { return j.path == name })

	// If the script was not found, return an error.
	if found == -1 {
		return ErrUnknownScript
	}
	// If the script is already stopped, return an error.
	if c.jobs[found].ID == 0 {
		return ErrAlreadyStopped
	}

	c.scheduler.Remove(c.jobs[found].ID)

	c.jobs[found].ID = 0

	return nil
}

// NewScriptController creates a new sensor controller for scripts.
func NewScriptsController(ctx context.Context, path string) (*Controller, error) {
	controller := &Controller{
		scheduler: cron.New(),
		logger:    logging.FromContext(ctx).With(slog.String("controller", "scripts")),
	}

	scripts, err := findScripts(path)
	if err != nil {
		return nil, fmt.Errorf("could not find scripts: %w", err)
	}

	controller.jobs = make([]job, 0, len(scripts))

	for _, s := range scripts {
		logAttrs := slog.Group("job", slog.String("script", s.path), slog.String("schedule", s.schedule))
		controller.jobs = append(controller.jobs, job{Script: *s, logAttrs: logAttrs})
	}

	return controller, nil
}

// findScripts locates scripts and returns a slice of scripts that the agent can
// run.
func findScripts(path string) ([]*Script, error) {
	var sensorScripts []*Script

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
