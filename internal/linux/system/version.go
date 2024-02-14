// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func Versions(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 3)
	defer close(sensorCh)
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Failed to retrieve host info.")
		close(sensorCh)
		return sensorCh
	}
	sensorCh <- &linux.Sensor{
		SensorTypeValue: linux.SensorKernel,
		Value:           info.KernelVersion,
		IsDiagnostic:    true,
		IconString:      "mdi:chip",
		SensorSrc:       linux.DataSrcProcfs,
	}
	sensorCh <- &linux.Sensor{
		SensorTypeValue: linux.SensorDistribution,
		Value:           cases.Title(language.English).String(info.Platform),
		IsDiagnostic:    true,
		IconString:      "mdi:linux",
		SensorSrc:       linux.DataSrcProcfs,
	}
	sensorCh <- &linux.Sensor{
		SensorTypeValue: linux.SensorVersion,
		Value:           info.PlatformVersion,
		IsDiagnostic:    true,
		IconString:      "mdi:numeric",
		SensorSrc:       linux.DataSrcProcfs,
	}
	return sensorCh
}
