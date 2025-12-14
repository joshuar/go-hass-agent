// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	abrtProblemsCheckInterval = 15 * time.Minute
	abrtProblemsCheckJitter   = time.Minute

	abrtProblemsPreferencesID = sensorsPrefPrefix + "abrt_problems"

	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

var (
	_ quartz.Job                  = (*problemsWorker)(nil)
	_ workers.PollingEntityWorker = (*problemsWorker)(nil)
)

type problemsWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	bus   *dbusx.Bus
	prefs *ProblemsPrefs
}

func NewProblemsWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &problemsWorker{
		WorkerMetadata:          models.SetWorkerMetadata("abrt", "ABRT Problems"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}
	// Check if we can fetch problem data, bail if we can't.
	_, err := dbusx.GetData[[]string](worker.bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
	if err != nil {
		return worker, fmt.Errorf("get abrt problems: %w", err)
	}

	defaultPrefs := &ProblemsPrefs{
		UpdateInterval: abrtProblemsCheckInterval.String(),
	}

	worker.prefs, err = workers.LoadWorkerPreferences(abrtProblemsPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = abrtProblemsCheckInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, abrtProblemsCheckJitter)

	return worker, nil
}

func (w *problemsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("schedule worker: %w", err)
	}
	return w.OutCh, nil
}

func (w *problemsWorker) Execute(ctx context.Context) error {
	// Get the list of problems.
	problems, err := w.getProblems()
	if err != nil {
		return fmt.Errorf("get abrt problems: %w", err)
	}
	w.OutCh <- w.generateEntity(ctx, problems)
	return nil
}

func (w *problemsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *problemsWorker) generateEntity(ctx context.Context, problems []string) models.Entity {
	problemDetails := make(map[string]map[string]any)
	// For each problem, fetch its details.
	for _, problem := range problems {
		details, err := w.getProblemDetails(problem)
		if err != nil {
			continue
		}

		problemDetails[problem] = parseProblem(details)
	}

	return sensor.NewSensor(ctx,
		sensor.WithName("Problems"),
		sensor.WithID("problems"),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits("problems"),
		sensor.WithIcon("mdi:alert"),
		sensor.WithState(len(problems)),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
		sensor.WithAttribute("problem_list", problemDetails),
	)
}

func (w *problemsWorker) getProblems() ([]string, error) {
	problems, err := dbusx.GetData[[]string](w.bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
	if err != nil {
		return nil, fmt.Errorf("error getting data: %w", err)
	}

	return problems, nil
}

func (w *problemsWorker) getProblemDetails(problem string) (map[string]string, error) {
	details, err := dbusx.GetData[map[string]string](w.bus,
		dBusProblemsDest,
		dBusProblemIntr,
		dBusProblemIntr+".GetInfo", problem, []string{"time", "count", "package", "reason"})

	switch {
	case err != nil:
		return nil, err
	default:
		return details, nil
	}
}

func parseProblem(details map[string]string) map[string]any {
	parsed := make(map[string]any)

	for key, value := range details {
		switch key {
		case "time":
			timeValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				parsed["time"] = 0
			} else {
				parsed["time"] = time.Unix(timeValue, 0).Format(time.RFC3339)
			}
		case "count":
			countValue, err := strconv.Atoi(value)
			if err != nil {
				parsed["count"] = 0
			} else {
				parsed["count"] = countValue
			}
		case "package", "reason":
			parsed[key] = value
		}
	}

	return parsed
}
