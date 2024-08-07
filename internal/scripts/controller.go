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
	Script
	ID cron.EntryID
}

type Controller struct {
	scheduler *cron.Cron
	logger    *slog.Logger
	jobs      []job
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

	script := c.jobs[found]

	sensorCh := make(chan sensor.Details)

	// Schedule the script.
	id, err := c.scheduler.AddFunc(script.Schedule(), func() {
		sensors, err := script.Execute()
		if err != nil {
			c.logger.Warn("Could not execute script.", slog.String("script", script.path), slog.Any("error", err))

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

	return nil
}

func (c *Controller) StartAll(_ context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	for _, job := range c.jobs {
		if job.ID > 0 {
			c.logger.Warn("Script already started.", slog.String("script", job.path))

			continue
		}

		// Create a closure to run the script on it's schedule.
		runFunc := func() {
			sensors, err := job.Script.Execute()
			if err != nil {
				c.logger.Warn("Could not execute script.", slog.String("script", job.path), slog.Any("error", err))

				return
			}

			for _, o := range sensors {
				sensorCh <- o
			}
		}
		// Add the script to the cron scheduler to run the closure on it's
		// defined schedule.
		id, err := c.scheduler.AddFunc(job.schedule, runFunc)
		if err != nil {
			c.logger.Warn("Unable to schedule script",
				slog.String("script", job.path),
				slog.String("schedule", job.schedule),
				slog.Any("error", err))
		} else {
			job.ID = id
			c.logger.Debug("Added cron job.",
				slog.String("script", job.path),
				slog.String("schedule", job.schedule),
				slog.Any("job_id", id))
		}
	}

	c.logger.Debug("Starting cron scheduler.")
	c.scheduler.Start()

	return sensorCh, nil
}

func (c *Controller) StopAll() error {
	for _, job := range c.jobs {
		c.logger.Debug("Removing cron job.", slog.String("script", job.path))
		c.scheduler.Remove(job.ID)
	}

	c.logger.Debug("Stopping cron scheduler.")
	waitCtx := c.scheduler.Stop()
	<-waitCtx.Done()

	c.logger.Debug("Exiting script controller.")

	return nil
}

// NewScriptController creates a new sensor controller for scripts.
func NewScriptsController(ctx context.Context, path string) (*Controller, error) {
	controller := &Controller{
		scheduler: cron.New(),
		logger:    logging.FromContext(ctx).WithGroup("scripts"),
	}

	scripts, err := findScripts(path)
	if err != nil {
		return nil, fmt.Errorf("could not find scripts: %w", err)
	}

	controller.jobs = make([]job, 0, len(scripts))
	for _, s := range scripts {
		controller.jobs = append(controller.jobs, job{Script: *s})
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
