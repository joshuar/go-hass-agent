// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package problems

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

type problemsSensor struct {
	list map[string]map[string]any
	linux.Sensor
}

func (s *problemsSensor) Attributes() any {
	return struct {
		ProblemList map[string]map[string]any `json:"Problem List"`
		DataSource  string                    `json:"Data Source"`
	}{
		ProblemList: s.list,
		DataSource:  linux.DataSrcDbus,
	}
}

func parseProblem(details map[string]string) map[string]any {
	parsed := make(map[string]any)
	for k, v := range details {
		switch k {
		case "time":
			timeValue, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Debug().Err(err).Msg("Could not parse problem time.")
				parsed["time"] = 0
			} else {
				parsed["time"] = time.Unix(timeValue, 0).Format(time.RFC3339)
			}
		case "count":
			countValue, err := strconv.Atoi(v)
			if err != nil {
				log.Debug().Err(err).Msg("Could not parse problem count.")
				parsed["count"] = 0
			} else {
				parsed["count"] = countValue
			}
		case "package", "reason":
			parsed[k] = v
		}
	}
	return parsed
}

type worker struct{}

func (w *worker) Interval() time.Duration { return 15 * time.Minute }

func (w *worker) Jitter() time.Duration { return time.Minute }

func (w *worker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	problems := &problemsSensor{
		list: make(map[string]map[string]any),
	}
	problems.SensorTypeValue = linux.SensorProblem
	problems.IconString = "mdi:alert"
	problems.UnitsString = "problems"
	problems.StateClassValue = types.StateClassMeasurement

	problemList, err := dbusx.GetData[[]string](ctx, dbusx.SystemBus, dBusProblemsDest, dBusProblemIntr, dBusProblemIntr+".GetProblems")
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve the list of ABRT problems: %w", err)
	}

	for _, p := range problemList {
		problemDetails, err := dbusx.GetData[map[string]string](ctx,
			dbusx.SystemBus,
			dBusProblemsDest,
			dBusProblemIntr,
			dBusProblemIntr+".GetInfo", p, []string{"time", "count", "package", "reason"})
		if problemDetails == nil || err != nil {
			log.Debug().Msg("No problems retrieved.")
		} else {
			problems.list[p] = parseProblem(problemDetails)
		}
	}
	if len(problems.list) > 0 {
		problems.Value = len(problems.list)
		return []sensor.Details{problems}, nil
	}
	return nil, nil
}

func NewProblemsWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "ABRT Problems Sensor",
			WorkerDesc: "Count of problems detected by ABRT (with details in sensor attributes).",
			Value:      &worker{},
		},
		nil
}
