// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strconv"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

type problemsSensor struct {
	list map[string]map[string]interface{}
	linuxSensor
}

func (s *problemsSensor) Attributes() interface{} {
	return struct {
		ProblemList map[string]map[string]interface{} `json:"Problem List"`
		DataSource  string                            `json:"Data Source"`
	}{
		ProblemList: s.list,
		DataSource:  "D-Bus",
	}
}

func parseProblem(details map[string]string) map[string]interface{} {
	parsed := make(map[string]interface{})
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
		case "package":
			fallthrough
		case "reason":
			parsed[k] = v
		}
	}
	return parsed
}

func ProblemsUpdater(ctx context.Context, status chan interface{}) {
	problems := func() {
		problems := &problemsSensor{
			list: make(map[string]map[string]interface{}),
		}
		problems.sensorType = problem
		problems.icon = "mdi:alert"
		problems.units = "problems"
		problems.stateClass = sensor.StateMeasurement

		problemList := NewBusRequest(SystemBus).
			Path(dBusProblemsDest).
			Destination(dBusProblemIntr).
			GetData(dBusProblemIntr + ".GetProblems").AsStringList()

		for _, p := range problemList {
			problemDetails := NewBusRequest(SystemBus).
				Path(dBusProblemsDest).
				Destination(dBusProblemIntr).
				GetData(dBusProblemIntr+".GetInfo", p, []string{"time", "count", "package", "reason"}).AsStringMap()
			if problemDetails == nil {
				log.Debug().Msg("No problems retrieved.")
			} else {
				problems.list[p] = parseProblem(problemDetails)
			}
		}
		if len(problems.list) > 0 {
			problems.value = len(problems.list)
			status <- problems
		}
	}

	helpers.PollSensors(ctx, problems, time.Minute*15, time.Minute)
}
