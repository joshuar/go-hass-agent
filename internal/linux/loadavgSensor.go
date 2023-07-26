// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/load"
)

type loadavg struct {
	load float64
	name sensorType
}

func (l *loadavg) Name() string {
	return l.name.String()
}

func (l *loadavg) ID() string {
	return strcase.ToSnake(l.name.String())
}

func (l *loadavg) Icon() string {
	return "mdi:chip"
}

func (l *loadavg) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (l *loadavg) DeviceClass() sensor.SensorDeviceClass {
	return 0
}

func (l *loadavg) StateClass() sensor.SensorStateClass {
	return sensor.StateMeasurement
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
	return struct {
		DataSource string `json:"Data Source"`
	}{
		DataSource: "procfs",
	}
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

	helpers.PollSensors(ctx, sendLoadAvgStats, time.Minute, time.Second*5)
}
