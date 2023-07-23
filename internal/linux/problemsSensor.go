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
	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
	"github.com/rs/zerolog/log"
)

const (
	dBusProblemsDest = "/org/freedesktop/problems"
	dBusProblemIntr  = "org.freedesktop.problems"
)

type problems struct {
	list map[string]map[string]interface{}
}

func (p *problems) Name() string {
	return "Problems"
}

func (p *problems) ID() string {
	return "problems"
}

func (p *problems) Icon() string {
	return "mdi:alert"
}

func (p *problems) SensorType() sensorType.SensorType {
	return sensorType.TypeSensor
}

func (p *problems) DeviceClass() deviceClass.SensorDeviceClass {
	return 0
}

func (p *problems) StateClass() stateClass.SensorStateClass {
	return stateClass.StateMeasurement
}

func (p *problems) State() interface{} {
	return len(p.list)
}

func (p *problems) Units() string {
	return "problems"
}

func (p *problems) Category() string {
	return ""
}

func (p *problems) Attributes() interface{} {
	return struct {
		ProblemList map[string]map[string]interface{} `json:"Problem List"`
		DataSource  string                            `json:"Data Source"`
	}{
		ProblemList: p.list,
		DataSource:  "D-Bus",
	}
}

func marshalProblemDetails(details map[string]string) map[string]interface{} {
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
		problems := &problems{
			list: make(map[string]map[string]interface{}),
		}

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
				problems.list[p] = marshalProblemDetails(problemDetails)
			}
		}
		if len(problems.list) > 0 {
			status <- problems
		}
	}

	helpers.PollSensors(ctx, problems, time.Minute*15, time.Minute)
}
