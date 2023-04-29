// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strconv"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/lthibault/jitterbug/v2"
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

func (p *problems) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (p *problems) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (p *problems) StateClass() hass.SensorStateClass {
	return hass.StateMeasurement
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
	return p.list
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

func sendAllProblems(deviceAPI *DeviceAPI, status chan interface{}) {
	problems := &problems{
		list: make(map[string]map[string]interface{}),
	}

	problemList := deviceAPI.SystemBusRequest().
		Path(dBusProblemsDest).
		Destination(dBusProblemIntr).
		GetData(dBusProblemIntr + ".GetProblems").AsStringList()

	for _, p := range problemList {
		problemDetails := deviceAPI.SystemBusRequest().
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

func ProblemsUpdater(ctx context.Context, status chan interface{}) {
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return
	}

	sendAllProblems(deviceAPI, status)

	ticker := jitterbug.New(
		time.Minute*15,
		&jitterbug.Norm{Stdev: time.Minute},
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Debug().Caller().Msg("Getting current problem list...")
				sendAllProblems(deviceAPI, status)
			}
		}
	}()

	// TODO: Turn into DBus watch

	// It would be better to watch for signals of problems being created or
	// removed. But currently it looks like only problem creation triggers a
	// signal.

	// problemWatch := &DBusWatchRequest{
	// 	bus:  systemBus,
	// 	path: "/org/freedesktop/Problems2",
	// 	match: []dbus.MatchOption{
	// 		dbus.WithMatchObjectPath("/org/freedesktop/Problems2"),
	// 		dbus.WithMatchInterface("org.freedesktop.Problems2"),
	// 	},
	// 	event: "org.freedesktop.Problems2.Crash",
	// 	eventHandler: func(s *dbus.Signal) {
	// 		spew.Dump(s)
	// 		sendAllProblems(deviceAPI, status)
	// 	},
	// }
	// deviceAPI.WatchEvents <- problemWatch

	// problemCatchAll := &DBusWatchRequest{
	// 	bus:  systemBus,
	// 	path: "/org/freedesktop/Problems2",
	// 	match: []dbus.MatchOption{
	// 		dbus.WithMatchObjectPath("/org/freedesktop/Problems2"),
	// 		dbus.WithMatchObjectPath("/org/freedesktop/problems"),
	// 	},
	// 	event: "org.freedesktop.DBus.Properties.PropertiesChanged",
	// 	eventHandler: func(s *dbus.Signal) {
	// 		spew.Dump(s)
	// 	},
	// }
	// deviceAPI.WatchEvents <- problemCatchAll
}
