// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package scripts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/id"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
)

var (
	ErrUnknownScript    = errors.New("unknown or nonexistent script")
	ErrAlreadyStarted   = errors.New("script already started")
	ErrAlreadyStopped   = errors.New("script already stopped")
	ErrSchedulingFailed = errors.New("failed to schedule script")
	ErrParseSchedule    = errors.New("could not parse script schedule")
)

const (
	scriptWorkerId   = "scripts"
	scriptWorkerDesc = "Custom script-based sensors"
)

// Worker represents the entity worker for handling scripts.
type Worker struct {
	logger  *slog.Logger
	scripts []*Script
	outCh   chan models.Entity
	*models.WorkerMetadata
}

// IsDisabled returns a boolean indicating whether the scripts worker is disabled.
func (c *Worker) IsDisabled() bool {
	return false
}

// States will execute all running scripts and returns their sensor entities.
func (c *Worker) States(ctx context.Context) []models.Entity {
	var allSensors []models.Entity

	for _, script := range c.scripts {
		scriptSensors, err := script.Run(ctx)
		if err != nil {
			c.logger.Warn("Could not retrieve script sensors",
				slog.String("script", script.Description()),
				slog.Any("error", err),
			)

			continue
		}

		allSensors = append(allSensors, scriptSensors...)
	}

	return allSensors
}

func (c *Worker) Start(ctx context.Context) (<-chan models.Entity, error) {
	c.outCh = make(chan models.Entity)

	scriptOutputs := make([]<-chan models.Entity, 0, len(c.scripts))

	for _, script := range c.scripts {
		// Parse the script cron schedule as a scheduler trigger.
		trigger, err := parseSchedule(script.Schedule())
		if err != nil {
			c.logger.Warn("Could not schedule script.",
				slog.String("script", script.Description()),
				slog.Any("error", err))

			continue
		}
		// Schedule the script.
		err = scheduler.Manager.ScheduleJob(id.ScriptJob, script, trigger)
		if err != nil {
			c.logger.Warn("Could not schedule script.",
				slog.String("script", script.Description()),
				slog.Any("error", err))

			continue
		}
		// Append to list of managed scripts.
		scriptOutputs = append(scriptOutputs, script.Start(ctx))
	}

	return mergeCh(ctx, scriptOutputs...), nil
}

func (c *Worker) Stop() error {
	close(c.outCh)

	return nil
}

// findScripts locates scripts and returns a slice of scripts that the agent can
// run.
func (c *Worker) findScripts(path string) ([]*Script, error) {
	var sensorScripts []*Script

	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not search for scripts: %w", err)
	}

	for _, scriptFile := range files {
		if isExecutable(scriptFile) {
			script, err := NewScript(scriptFile)
			if err != nil {
				c.logger.Warn("Script error.",
					slog.Any("error", err),
				)

				continue
			}

			sensorScripts = append(sensorScripts, script)
		}
	}

	return sensorScripts, nil
}

// NewScriptController creates a new sensor worker for scripts.
func NewScriptsWorker(ctx context.Context) (*Worker, error) {
	scriptPath := filepath.Join(preferences.PathFromCtx(ctx), "scripts")

	worker := &Worker{
		WorkerMetadata: models.SetWorkerMetadata(scriptWorkerId, scriptWorkerDesc),
		logger:         logging.FromContext(ctx).WithGroup("scripts"),
	}

	scripts, err := worker.findScripts(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("could not find scripts: %w", err)
	}

	worker.scripts = scripts

	return worker, nil
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
