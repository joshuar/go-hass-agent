// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/disk"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type diskSensor struct {
	stats *disk.UsageStat
	linux.Sensor
}

func newDiskSensor(d *disk.UsageStat) *diskSensor {
	s := &diskSensor{}
	s.IconString = "mdi:harddisk"
	s.StateClassValue = types.StateClassTotal
	s.UnitsString = "%"
	s.stats = d
	s.Value = math.Round(d.UsedPercent/0.05) * 0.05
	return s
}

// diskUsageState implements hass.SensorUpdate

func (d *diskSensor) Name() string {
	return "Mountpoint " + d.stats.Path + " Usage"
}

func (d *diskSensor) ID() string {
	if d.stats.Path == "/" {
		return "mountpoint_root"
	}
	return "mountpoint" + strings.ReplaceAll(d.stats.Path, "/", "_")
}

func (d *diskSensor) Attributes() any {
	return struct {
		DataSource string `json:"Data Source"`
		Stats      disk.UsageStat
	}{
		DataSource: linux.DataSrcProcfs,
		Stats:      *d.stats,
	}
}

func UsageUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 1)
	sendDiskUsageStats := func(_ time.Duration) {
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
			sensorCh <- newDiskSensor(usage)
		}
	}

	go helpers.PollSensors(ctx, sendDiskUsageStats, time.Minute, time.Second*5)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped disk usage sensors.")
	}()
	return sensorCh
}
