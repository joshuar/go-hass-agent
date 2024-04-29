// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func Versions(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Failed to retrieve host info.")
		close(sensorCh)
		return sensorCh
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sensorCh <- &linux.Sensor{
			SensorTypeValue: linux.SensorKernel,
			Value:           info.KernelVersion,
			IsDiagnostic:    true,
			IconString:      "mdi:chip",
			SensorSrc:       linux.DataSrcProcfs,
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sensorCh <- &linux.Sensor{
			SensorTypeValue: linux.SensorDistribution,
			Value:           cases.Title(language.English).String(info.Platform),
			IsDiagnostic:    true,
			IconString:      "mdi:linux",
			SensorSrc:       linux.DataSrcProcfs,
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sensorCh <- &linux.Sensor{
			SensorTypeValue: linux.SensorVersion,
			Value:           info.PlatformVersion,
			IsDiagnostic:    true,
			IconString:      "mdi:numeric",
			SensorSrc:       linux.DataSrcProcfs,
		}
	}()
	go func() {
		defer close(sensorCh)
		wg.Wait()
	}()
	return sensorCh
}
