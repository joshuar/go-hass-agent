// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

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
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

type Worker struct {
	sensorCh  chan sensor.Entity
	scheduler *cron.Cron
	logger    *slog.Logger
	jobs      []job
}

// TODO: implement ability to disable.
func (c *Worker) Disabled() bool {
	return false
}

func (c *Worker) ID() string {
	return "scripts"
}

func (c *Worker) States(_ context.Context) []sensor.Entity {
	var sensors []sensor.Entity

	for _, worker := range c.activeJobs() {
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

func (c *Worker) activeJobs() []string {
	var activeScripts []string

	for _, job := range c.jobs {
		if job.ID != 0 {
			activeScripts = append(activeScripts, job.path)
		}
	}

	return activeScripts
}

func (c *Worker) inactiveJobs() []string {
	var inactiveScripts []string

	for _, job := range c.jobs {
		if job.ID == 0 {
			inactiveScripts = append(inactiveScripts, job.path)
		}
	}

	return inactiveScripts
}

func (c *Worker) Start(_ context.Context) (<-chan sensor.Entity, error) {
	c.sensorCh = make(chan sensor.Entity)

	for idx := range c.inactiveJobs() {
		// Schedule the script.
		id, err := c.scheduler.AddFunc(c.jobs[idx].Schedule(), func() {
			sensors, err := c.jobs[idx].Execute()
			if err != nil {
				c.logger.Warn("Could not execute script.",
					c.jobs[idx].logAttrs,
					slog.Any("error", err))

				return
			}

			for _, o := range sensors {
				c.sensorCh <- o
			}
		})
		if err != nil {
			c.logger.Warn("Could not schedule script.",
				c.jobs[idx].logAttrs,
				slog.Any("error", err))

			continue
		}

		// Update the job id.
		c.jobs[idx].ID = id

		c.logger.Debug("Scheduled script.",
			c.jobs[idx].logAttrs)
	}

	// Start the scheduler.
	c.scheduler.Start()

	// Return the new sensor channel for the script.
	return c.sensorCh, nil
}

func (c *Worker) Stop() error {
	for idx := range c.activeJobs() {
		// If the script is already stopped, return an error.
		if c.jobs[idx].ID == 0 {
			c.logger.Warn("Script already stopped.",
				c.jobs[idx].logAttrs)

			continue
		}

		c.scheduler.Remove(c.jobs[idx].ID)

		c.jobs[idx].ID = 0
	}

	// Stop the scheduler.
	c.scheduler.Stop()

	close(c.sensorCh)

	return nil
}

// NewScriptController creates a new sensor worker for scripts.
func NewScriptsWorker(ctx context.Context) (*Worker, error) {
	scriptPath := filepath.Join(preferences.PathFromCtx(ctx), "scripts")

	worker := &Worker{
		scheduler: cron.New(),
		logger:    logging.FromContext(ctx).With(slog.String("controller", "scripts")),
	}

	scripts, err := findScripts(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("could not find scripts: %w", err)
	}

	worker.jobs = make([]job, 0, len(scripts))

	for _, s := range scripts {
		logAttrs := slog.Group("job", slog.String("script", s.path), slog.String("schedule", s.schedule))
		worker.jobs = append(worker.jobs, job{Script: *s, logAttrs: logAttrs})
	}

	return worker, nil
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
