// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/disk"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type diskUsageSensor struct {
	stats *disk.UsageStat
	linux.Sensor
}

func newDiskUsageSensor(d *disk.UsageStat) *diskUsageSensor {
	s := &diskUsageSensor{}
	s.IconString = "mdi:harddisk"
	s.StateClassValue = types.StateClassTotal
	s.UnitsString = "%"
	s.stats = d
	s.Value = math.Round(d.UsedPercent/0.05) * 0.05
	return s
}

// diskUsageState implements hass.SensorUpdate

func (d *diskUsageSensor) Name() string {
	return "Mountpoint " + d.stats.Path + " Usage"
}

func (d *diskUsageSensor) ID() string {
	if d.stats.Path == "/" {
		return "mountpoint_root"
	}
	return "mountpoint" + strings.ReplaceAll(d.stats.Path, "/", "_")
}

func (d *diskUsageSensor) Attributes() any {
	return struct {
		DataSource string `json:"Data Source"`
		Stats      disk.UsageStat
	}{
		DataSource: linux.DataSrcProcfs,
		Stats:      *d.stats,
	}
}

type usageWorker struct{}

func (w *usageWorker) Interval() time.Duration { return time.Minute }

func (w *usageWorker) Jitter() time.Duration { return 10 * time.Second }

func (w *usageWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	var sensors []sensor.Details
	p, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of physical partitions: %w", err)
	}
	for _, partition := range p {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to get usage info for mountpount %s.", partition.Mountpoint)
			continue
		}
		sensors = append(sensors, newDiskUsageSensor(usage))
	}
	return sensors, nil
}

func NewUsageWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Disk Usage Sensors",
			WorkerDesc: "Disk Space Usage.",
			Value:      &usageWorker{},
		},
		nil
}
