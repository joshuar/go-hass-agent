// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package problems

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	problemInterval = 15 * time.Minute
	problemJitter   = time.Minute

	problemsWorkerID = "abrt_problems_sensor"

	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

var ErrNoProblemDetails = errors.New("no details found")

type problemsSensor struct {
	list map[string]map[string]any
	linux.Sensor
}

func (s *problemsSensor) Attributes() map[string]any {
	attributes := s.Sensor.Attributes()

	if s.list != nil {
		attributes["problem_list"] = s.list
	}

	return attributes
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

type worker struct {
	getProblems       func() ([]string, error)
	getProblemDetails func(problem string) (map[string]string, error)
	bus               *dbusx.Bus
}

func (w *worker) Interval() time.Duration { return problemInterval }

func (w *worker) Jitter() time.Duration { return problemJitter }

func (w *worker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	problemsSensor := &problemsSensor{
		list: make(map[string]map[string]any),
		Sensor: linux.Sensor{
			DisplayName:     "Problems",
			UniqueID:        "problems",
			IconString:      "mdi:alert",
			UnitsString:     "problems",
			StateClassValue: types.StateClassMeasurement,
			DataSource:      linux.DataSrcDbus,
		},
	}

	// Get the list of problems.
	problems, err := w.getProblems()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of problems from D-Bus: %w", err)
	}
	// Set the value of the sensor to the total count of problems.
	problemsSensor.Value = len(problems)
	// For each problem, fetch its details.
	for _, problem := range problems {
		details, err := w.getProblemDetails(problem)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", problemsWorkerID)).
				Debug("Unable to get problem details.",
					slog.String("problem", problem),
					slog.Any("error", err))
		} else {
			problemsSensor.list[problem] = parseProblem(details)
		}
	}

	return []sensor.Details{problemsSensor}, nil
}

func NewProblemsWorker(ctx context.Context) (*linux.SensorWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	// Check if we can fetch problem data, bail if we can't.
	_, err := dbusx.GetData[[]string](bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch ABRT problems from D-Bus: %w", err)
	}

	return &linux.SensorWorker{
			Value: &worker{
				getProblems: func() ([]string, error) {
					problems, err := dbusx.GetData[[]string](bus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
					if err != nil {
						return nil, fmt.Errorf("error getting data: %w", err)
					}

					return problems, nil
				},
				getProblemDetails: func(problem string) (map[string]string, error) {
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
				},
				bus: bus,
			},
			WorkerID: problemsWorkerID,
		},
		nil
}
