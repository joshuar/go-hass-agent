// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/load"
)

const (
	load1 loadavgStat = iota + 1
	load5
	load15
)

type loadavgStat int

type loadavg struct {
	load float64
	name loadavgStat
}

func (l *loadavg) Name() string {
	switch l.name {
	case load1:
		return "CPU load average (1 min)"
	case load5:
		return "CPU load average (5 min)"
	case load15:
		return "CPU load average (15 min)"
	default:
		return "CPU Load Average"
	}
}

func (l *loadavg) ID() string {
	switch l.name {
	case load1:
		return "cpu_load_avg_1_min"
	case load5:
		return "cpu_load_avg_5_min"
	case load15:
		return "cpu_load_avg_15_min"
	default:
		return "cpu_load_avg"
	}
}

func (l *loadavg) Icon() string {
	return "mdi:chip"
}

func (l *loadavg) SensorType() sensorType.SensorType {
	return sensorType.TypeSensor
}

func (l *loadavg) DeviceClass() deviceClass.SensorDeviceClass {
	return 0
}

func (l *loadavg) StateClass() stateClass.SensorStateClass {
	return stateClass.StateMeasurement
}

func (l *loadavg) State() interface{} {
	return l.load
}

func (l *loadavg) Units() string {
	return "load"
}

func (l *loadavg) Category() string {
	return ""
}

func (l *loadavg) Attributes() interface{} {
	return nil
}

func LoadAvgUpdater(ctx context.Context, status chan interface{}) {

	sendLoadAvgStats := func() {
		var latest *load.AvgStat
		var err error
		if latest, err = load.AvgWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching loadavg stats.")
			return
		}
		status <- &loadavg{
			load: latest.Load1,
			name: load1,
		}
		status <- &loadavg{
			load: latest.Load5,
			name: load5,
		}
		status <- &loadavg{
			load: latest.Load15,
			name: load15,
		}
	}

	pollSensors(ctx, sendLoadAvgStats, time.Minute, time.Second*5)
}
