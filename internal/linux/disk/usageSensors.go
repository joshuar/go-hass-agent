// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"math"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type diskUsageSensor struct {
	stats *disk.UsageStat
	linux.Sensor
}

//nolint:mnd
func newDiskUsageSensor(stat *disk.UsageStat) *diskUsageSensor {
	return &diskUsageSensor{
		Sensor: linux.Sensor{
			IconString:      "mdi:harddisk",
			StateClassValue: types.StateClassTotal,
			UnitsString:     "%",
			Value:           math.Round(stat.UsedPercent/0.05) * 0.05,
		},
		stats: stat,
	}
}

func (d *diskUsageSensor) Name() string {
	return "Mountpoint " + d.stats.Path + " Usage"
}

func (d *diskUsageSensor) ID() string {
	if d.stats.Path == "/" {
		return "mountpoint_root"
	}

	return "mountpoint" + strings.ReplaceAll(d.stats.Path, "/", "_")
}

func (d *diskUsageSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["data_source"] = linux.DataSrcProcfs
	attributes["stats"] = d.stats

	return attributes
}
