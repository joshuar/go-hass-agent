// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Versions(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 3)
	defer close(sensorCh)
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Failed to retrieve host info.")
		close(sensorCh)
		return sensorCh
	}
	sensorCh <- &linuxSensor{
		sensorType:   kernel,
		value:        info.KernelVersion,
		isDiagnostic: true,
		icon:         "mdi:chip",
		source:       srcProcfs,
	}
	sensorCh <- &linuxSensor{
		sensorType:   distribution,
		value:        cases.Title(language.English).String(info.Platform),
		isDiagnostic: true,
		icon:         "mdi:linux",
		source:       srcProcfs,
	}
	sensorCh <- &linuxSensor{
		sensorType:   version,
		value:        info.PlatformVersion,
		isDiagnostic: true,
		icon:         "mdi:numeric",
		source:       srcProcfs,
	}
	return sensorCh
}
