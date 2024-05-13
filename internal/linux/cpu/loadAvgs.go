// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cpu

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/load"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type loadavgSensor struct {
	linux.Sensor
}

func LoadAvgUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	sendLoadAvgStats := func(_ time.Duration) {
		var latest *load.AvgStat
		var err error
		if latest, err = load.AvgWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching loadavg stats.")
			return
		}
		for _, loadType := range []linux.SensorTypeValue{linux.SensorLoad1, linux.SensorLoad5, linux.SensorLoad15} {
			l := &loadavgSensor{}
			l.IconString = "mdi:chip"
			l.UnitsString = "load"
			l.SensorSrc = linux.DataSrcProcfs
			l.StateClassValue = types.StateClassMeasurement
			switch loadType {
			case linux.SensorLoad1:
				l.Value = latest.Load1
				l.SensorTypeValue = linux.SensorLoad1
			case linux.SensorLoad5:
				l.Value = latest.Load5
				l.SensorTypeValue = linux.SensorLoad5
			case linux.SensorLoad15:
				l.Value = latest.Load15
				l.SensorTypeValue = linux.SensorLoad15
			}
			go func() {
				sensorCh <- l
			}()
		}
	}

	// Send load averages on start up.
	go func() {
		sendLoadAvgStats(0)
	}()

	go helpers.PollSensors(ctx, sendLoadAvgStats, time.Minute, time.Second*5)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped load average sensors.")
	}()
	return sensorCh
}
