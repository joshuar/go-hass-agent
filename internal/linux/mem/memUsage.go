// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mem

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/mem"
)

type memorySensor struct {
	linux.Sensor
}

func (s *memorySensor) Attributes() any {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
	}{
		NativeUnit: s.UnitsString,
		DataSource: s.SensorSrc,
	}
}

func Updater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 5)
	sendMemStats := func(_ time.Duration) {
		stats := []linux.SensorTypeValue{linux.SensorMemTotal, linux.SensorMemAvail, linux.SensorMemUsed, linux.SensorSwapTotal, linux.SensorSwapFree}
		var memDetails *mem.VirtualMemoryStat
		var err error
		if memDetails, err = mem.VirtualMemoryWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching memory stats.")
			return
		}
		for _, stat := range stats {
			var statValue uint64
			switch stat {
			case linux.SensorMemTotal:
				statValue = memDetails.Total
			case linux.SensorMemAvail:
				statValue = memDetails.Available
			case linux.SensorMemUsed:
				statValue = memDetails.Used
			case linux.SensorSwapTotal:
				statValue = memDetails.SwapTotal
			case linux.SensorSwapFree:
				statValue = memDetails.SwapFree
				// case UsedSwapMemory:
				// 	return m.memStats.SwapCached
			}
			state := &memorySensor{
				linux.Sensor{
					Value:            statValue,
					SensorTypeValue:  stat,
					IconString:       "mdi:memory",
					UnitsString:      "B",
					SensorSrc:        linux.DataSrcProcfs,
					DeviceClassValue: sensor.Data_size,
					StateClassValue:  sensor.StateTotal,
				},
			}
			sensorCh <- state
		}
	}

	go helpers.PollSensors(ctx, sendMemStats, time.Minute, time.Second*5)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped memory usage sensors.")
	}()
	return sensorCh
}
