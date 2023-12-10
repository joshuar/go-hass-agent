// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/load"
)

type loadavgSensor struct {
	linuxSensor
}

func LoadAvgUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 3)
	sendLoadAvgStats := func(_ time.Duration) {
		var latest *load.AvgStat
		var err error
		if latest, err = load.AvgWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching loadavg stats.")
			return
		}
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
			sensorCh <- l
		}
	}

	go helpers.PollSensors(ctx, sendLoadAvgStats, time.Minute, time.Second*5)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped load average sensors.")
	}()
	return sensorCh
}
