// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/disk"
)

type diskSensor struct {
	stats *disk.UsageStat
	linuxSensor
}

func newDiskSensor(d *disk.UsageStat) *diskSensor {
	s := &diskSensor{}
	s.icon = "mdi:harddisk"
	s.stateClass = sensor.StateTotal
	s.units = "%"
	s.stats = d
	s.value = math.Round(d.UsedPercent/0.05) * 0.05
	return s
}

// diskUsageState implements hass.SensorUpdate

func (d *diskSensor) Name() string {
	return "Mountpoint " + d.stats.Path + " Usage"
}

func (d *diskSensor) ID() string {
	if d.stats.Path == "/" {
		return "mountpoint_root"
	} else {
		return "mountpoint" + strings.ReplaceAll(d.stats.Path, "/", "_")
	}
}

func (s *diskSensor) Attributes() interface{} {
	return struct {
		DataSource string `json:"Data Source"`
		Stats      disk.UsageStat
	}{
		DataSource: "SOURCE_PROCFS",
		Stats:      *s.stats,
	}
}

func DiskUsageUpdater(ctx context.Context, status chan interface{}) {

	sendDiskUsageStats := func() {
		p, err := disk.PartitionsWithContext(ctx, false)
		if err != nil {
			log.Debug().Err(err).
				Msg("Could not retrieve list of physical partitions.")
			return
		}
		for _, partition := range p {
			usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
			if err != nil {
				log.Debug().Err(err).
					Msgf("Failed to get usage info for mountpount %s.", partition.Mountpoint)
				return
			}
			status <- newDiskSensor(usage)
		}
	}

	helpers.PollSensors(ctx, sendDiskUsageStats, time.Minute, time.Second*5)
}
