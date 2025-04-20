// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	abrtProblemsCheckInterval = 15 * time.Minute
	abrtProblemsCheckJitter   = time.Minute

	abrtProblemsWorkerID      = "abrt_problems_sensor"
	abrtProblemsWorkerDesc    = "ABRT problems"
	abrtProblemsPreferencesID = sensorsPrefPrefix + "abrt_problems"

	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

var (
	_ quartz.Job                  = (*problemsWorker)(nil)
	_ workers.PollingEntityWorker = (*problemsWorker)(nil)
)

var (
	ErrNewABRTSensor    = errors.New("could not create ABRT sensor")
	ErrInitABRTWorker   = errors.New("could not init ABRT worker")
	ErrNoProblemDetails = errors.New("no details found")
)

type problemsWorker struct {
	bus   *dbusx.Bus
	prefs *ProblemsPrefs
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
}

func (w *problemsWorker) generateEntity(ctx context.Context, problems []string) (*models.Entity, error) {
	problemDetails := make(map[string]map[string]any)
	// For each problem, fetch its details.
	for _, problem := range problems {
		details, err := w.getProblemDetails(problem)
		if err != nil {
			continue
		}

		problemDetails[problem] = parseProblem(details)
	}

	abrtSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Problems"),
		sensor.WithID("problems"),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits("problems"),
		sensor.WithIcon("mdi:alert"),
		sensor.WithState(len(problems)),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
		sensor.WithAttribute("problem_list", problemDetails),
	)
	if err != nil {
		return nil, errors.Join(ErrNewABRTSensor, err)
	}

	return &abrtSensor, nil
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
	case details == nil:
		return nil, ErrNoProblemDetails
	case err != nil:
		return nil, err
	default:
		return details, nil
	}
}

func (w *problemsWorker) Execute(ctx context.Context) error {
	// Get the list of problems.
	problems, err := w.getProblems()
	if err != nil {
		return fmt.Errorf("could not retrieve list of problems from D-Bus: %w", err)
	}

	entity, err := w.generateEntity(ctx, problems)
	if err != nil {
		return fmt.Errorf("could not generate problem sensor: %w", err)
	}
	w.OutCh <- *entity
	return nil
}

func (w *problemsWorker) PreferencesID() string {
	return abrtProblemsPreferencesID
}

func (w *problemsWorker) DefaultPreferences() ProblemsPrefs {
	return ProblemsPrefs{
		UpdateInterval: abrtProblemsCheckInterval.String(),
	}
}

func (w *problemsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *problemsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

func NewProblemsWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}
	// Check if we can fetch problem data, bail if we can't.
	if _, err := dbusx.GetData[[]string](bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems"); err != nil {
		return nil, errors.Join(ErrInitABRTWorker,
			fmt.Errorf("unable to fetch ABRT problems from D-Bus: %w", err))
	}

	worker := &problemsWorker{
		bus:                     bus,
		WorkerMetadata:          models.SetWorkerMetadata(abrtProblemsWorkerID, abrtProblemsWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitABRTWorker, err)
	}
	worker.prefs = prefs

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", abrtProblemsWorkerID),
			slog.String("given_interval", worker.prefs.UpdateInterval),
			slog.String("default_interval", abrtProblemsCheckInterval.String()))

		pollInterval = abrtProblemsCheckInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, abrtProblemsCheckJitter)

	return worker, nil
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
