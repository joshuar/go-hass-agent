// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/load"
)

type loadavgSensor struct {
	linuxSensor
}

func LoadAvgUpdater(ctx context.Context, tracker device.SensorTracker) {
	sendLoadAvgStats := func() {
		var latest *load.AvgStat
		var err error
		if latest, err = load.AvgWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching loadavg stats.")
			return
		}
		var sensors []interface{}
		for _, loadType := range []sensorType{load1, load5, load15} {
			l := &loadavgSensor{}
			l.icon = "mdi:chip"
			l.units = "load"
			l.source = srcProcfs
			l.stateClass = sensor.StateMeasurement
			switch loadType {
			case load1:
				l.value = latest.Load1
				l.sensorType = load1
			case load5:
				l.value = latest.Load5
				l.sensorType = load5
			case load15:
				l.value = latest.Load15
				l.sensorType = load15
			}
			sensors = append(sensors, l)
		}
		if err := tracker.UpdateSensors(ctx, sensors...); err != nil {
			log.Error().Err(err).Msg("Could not update load average sensors.")
		}
	}

	helpers.PollSensors(ctx, sendLoadAvgStats, time.Minute, time.Second*5)
}
