// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cpu

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type cpuUsageSensor struct {
	linux.Sensor
}

func UsageUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	sendCPUUsage := func(d time.Duration) {
		usage, err := cpu.Percent(d, false)
		if err != nil {
			log.Warn().Err(err).Msg("Could not retrieve CPU usage.")
		}
		s := &cpuUsageSensor{}
		s.IconString = "mdi:chip"
		s.UnitsString = "%"
		s.SensorSrc = linux.DataSrcProcfs
		s.StateClassValue = types.StateClassMeasurement
		s.Value = usage[0]
		s.SensorTypeValue = linux.SensorCPUPc

		go func() {
			sensorCh <- s
		}()
	}

	// Send CPU usage on start.
	go func() {
		sendCPUUsage(0)
	}()

	go helpers.PollSensors(ctx, sendCPUUsage, time.Second*10, time.Second)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped CPU usage sensor.")
	}()
	return sensorCh
}
