// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mem

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

var stats = []linux.SensorTypeValue{
	linux.SensorMemTotal,
	linux.SensorMemAvail,
	linux.SensorMemUsed,
	linux.SensorMemPc,
	linux.SensorSwapTotal,
	linux.SensorSwapFree,
	linux.SensorSwapPc,
}

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

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 5)
	sendMemStats := func(_ time.Duration) {
		var memDetails *mem.VirtualMemoryStat
		var err error
		if memDetails, err = mem.VirtualMemoryWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching memory stats.")
			return
		}
		for _, stat := range stats {
			value, unit, deviceClass, stateClass := parseSensorType(stat, memDetails)
			state := &memorySensor{
				linux.Sensor{
					Value:            value,
					SensorTypeValue:  stat,
					IconString:       "mdi:memory",
					UnitsString:      unit,
					SensorSrc:        linux.DataSrcProcfs,
					DeviceClassValue: deviceClass,
					StateClassValue:  stateClass,
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

func parseSensorType(t linux.SensorTypeValue, d *mem.VirtualMemoryStat) (value any, unit string, deviceClass types.DeviceClass, stateClass types.StateClass) {
	switch t {
	case linux.SensorMemTotal:
		return d.Total, "B", types.DeviceClassDataSize, types.StateClassTotal
	case linux.SensorMemAvail:
		return d.Available, "B", types.DeviceClassDataSize, types.StateClassTotal
	case linux.SensorMemUsed:
		return d.Used, "B", types.DeviceClassDataSize, types.StateClassTotal
	case linux.SensorMemPc:
		return float64(d.Used) / float64(d.Total) * 100, "%", 0, types.StateClassMeasurement
	case linux.SensorSwapTotal:
		return d.SwapTotal, "B", types.DeviceClassDataSize, types.StateClassTotal
	case linux.SensorSwapFree:
		return d.SwapFree, "B", types.DeviceClassDataSize, types.StateClassTotal
	case linux.SensorSwapPc:
		return float64(d.SwapCached) / float64(d.SwapTotal) * 100, "%", 0, types.StateClassMeasurement
	default:
		return sensor.StateUnknown, "", 0, 0
	}
}
