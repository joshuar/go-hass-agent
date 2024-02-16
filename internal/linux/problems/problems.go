// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package problems

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
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

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 1)
	problems := func(_ time.Duration) {
		problems := &problemsSensor{
			list: make(map[string]map[string]any),
		}
		problems.SensorTypeValue = linux.SensorProblem
		problems.IconString = "mdi:alert"
		problems.UnitsString = "problems"
		problems.StateClassValue = types.StateClassMeasurement

		req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
			Path(dBusProblemsDest).
			Destination(dBusProblemIntr)

		problemList, err := dbusx.GetData[[]string](req, dBusProblemIntr+".GetProblems")
		if err != nil {
			log.Warn().Err(err).Msg("Could not retrieve problem list.")
		}

		for _, p := range problemList {
			problemDetails, err := dbusx.GetData[map[string]string](req, dBusProblemIntr+".GetInfo", p, []string{"time", "count", "package", "reason"})
			if problemDetails == nil || err != nil {
				log.Debug().Msg("No problems retrieved.")
			} else {
				problems.list[p] = parseProblem(problemDetails)
			}
		}
		if len(problems.list) > 0 {
			problems.Value = len(problems.list)
			sensorCh <- problems
		}
	}

	go helpers.PollSensors(ctx, problems, time.Minute*15, time.Minute)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped problems sensor.")
	}()
	return sensorCh
}
