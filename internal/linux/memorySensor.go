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
	"github.com/shirou/gopsutil/v3/mem"
)

type memorySensor struct {
	linuxSensor
}

func (s *memorySensor) Attributes() interface{} {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
	}{
		NativeUnit: s.units,
		DataSource: s.source,
	}
}

func MemoryUpdater(ctx context.Context, tracker device.SensorTracker) {
	sendMemStats := func() {
		stats := []sensorType{memTotal, memAvail, memUsed, swapTotal, swapFree}
		var memDetails *mem.VirtualMemoryStat
		var err error
		if memDetails, err = mem.VirtualMemoryWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching memory stats.")
			return
		}
		var sensors []interface{}
		for _, stat := range stats {
			var statValue uint64
			switch stat {
			case memTotal:
				statValue = memDetails.Total
			case memAvail:
				statValue = memDetails.Available
			case memUsed:
				statValue = memDetails.Used
			case swapTotal:
				statValue = memDetails.SwapTotal
			case swapFree:
				statValue = memDetails.SwapFree
				// case UsedSwapMemory:
				// 	return m.memStats.SwapCached
			}
			state := &memorySensor{
				linuxSensor{
					value:       statValue,
					sensorType:  stat,
					icon:        "mdi:memory",
					units:       "B",
					source:      srcProcfs,
					deviceClass: sensor.Data_size,
					stateClass:  sensor.StateTotal,
				},
			}
			sensors = append(sensors, state)
		}
		if err := tracker.UpdateSensors(ctx, sensors...); err != nil {
			log.Error().Err(err).Msg("Could not update memory sensors.")
		}
	}

	helpers.PollSensors(ctx, sendMemStats, time.Minute, time.Second*5)
}
