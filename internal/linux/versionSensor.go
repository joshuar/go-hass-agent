// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Versions(ctx context.Context, tracker device.SensorTracker) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Failed to retrieve host info.")
	}
	var sensors []interface{}
	sensors = append(sensors,
		&linuxSensor{
			sensorType: kernel,
			value:      info.KernelVersion,
			diagnostic: true,
			icon:       "mdi:chip",
			source:     "SOURCE_PROCFS",
		}, &linuxSensor{
			sensorType: distribution,
			value:      cases.Title(language.English).String(info.Platform),
			diagnostic: true,
			icon:       "mdi:linux",
			source:     "SOURCE_PROCFS",
		},
		&linuxSensor{
			sensorType: version,
			value:      info.PlatformVersion,
			diagnostic: true,
			icon:       "mdi:numeric",
			source:     "SOURCE_PROCFS",
		},
	)
	tracker.UpdateSensors(ctx, sensors...)
}
