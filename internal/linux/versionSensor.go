// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Versions(ctx context.Context, status chan interface{}) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Failed to retrieve host info.")
	}
	status <- &linuxSensor{
		sensorType: kernel,
		value:      info.KernelVersion,
		diagnostic: true,
		icon:       "mdi:chip",
		source:     "procfs",
	}
	status <- &linuxSensor{
		sensorType: distribution,
		value:      cases.Title(language.English).String(info.Platform),
		diagnostic: true,
		icon:       "mdi:linux",
		source:     "procfs",
	}
	status <- &linuxSensor{
		sensorType: version,
		value:      info.PlatformVersion,
		diagnostic: true,
		icon:       "mdi:numeric",
		source:     "procfs",
	}
}
