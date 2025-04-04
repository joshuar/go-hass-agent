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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	abrtProblemsCheckInterval = 15 * time.Minute
	abrtProblemsCheckJitter   = time.Minute

	abrtProblemsWorkerID      = "abrt_problems_sensor"
	abrtProblemsPreferencesID = sensorsPrefPrefix + "abrt_problems"

	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

var (
	ErrNewABRTSensor    = errors.New("could not create ABRT sensor")
	ErrInitABRTWorker   = errors.New("could not init ABRT worker")
	ErrNoProblemDetails = errors.New("no details found")
)

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

func (w *problemsWorker) newProblemsSensor(ctx context.Context, problems []string) (*models.Entity, error) {
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

type problemsWorker struct {
	getProblems       func() ([]string, error)
	getProblemDetails func(problem string) (map[string]string, error)
	bus               *dbusx.Bus
	prefs             *ProblemsPrefs
}

func (w *problemsWorker) UpdateDelta(_ time.Duration) {}

func (w *problemsWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	// Get the list of problems.
	problems, err := w.getProblems()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of problems from D-Bus: %w", err)
	}

	entity, err := w.newProblemsSensor(ctx, problems)
	if err != nil {
		return nil, fmt.Errorf("could not generate problem sensor: %w", err)
	}

	return []models.Entity{*entity}, nil
}

func (w *problemsWorker) PreferencesID() string {
	return abrtProblemsPreferencesID
}

func (w *problemsWorker) DefaultPreferences() ProblemsPrefs {
	return ProblemsPrefs{
		UpdateInterval: abrtProblemsCheckInterval.String(),
	}
}

func NewProblemsWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	var err error

	problemsWorker := &problemsWorker{}

	problemsWorker.prefs, err = preferences.LoadWorker(problemsWorker)
	if err != nil {
		return nil, errors.Join(ErrInitABRTWorker, err)
	}

	//nolint:nilnil
	if problemsWorker.prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(problemsWorker.prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", abrtProblemsWorkerID),
			slog.String("given_interval", problemsWorker.prefs.UpdateInterval),
			slog.String("default_interval", abrtProblemsCheckInterval.String()))

		pollInterval = abrtProblemsCheckInterval
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	problemsWorker.bus = bus

	// Check if we can fetch problem data, bail if we can't.
	_, err = dbusx.GetData[[]string](bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
	if err != nil {
		return nil, errors.Join(ErrInitABRTWorker,
			fmt.Errorf("unable to fetch ABRT problems from D-Bus: %w", err))
	}

	problemsWorker.getProblems = func() ([]string, error) {
		problems, err := dbusx.GetData[[]string](bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
		if err != nil {
			return nil, fmt.Errorf("error getting data: %w", err)
		}

		return problems, nil
	}

	problemsWorker.getProblemDetails = func(problem string) (map[string]string, error) {
		details, err := dbusx.GetData[map[string]string](bus,
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

	worker := linux.NewPollingSensorWorker(abrtProblemsWorkerID, pollInterval, abrtProblemsCheckJitter)
	worker.PollingSensorType = problemsWorker

	return worker, nil
}
