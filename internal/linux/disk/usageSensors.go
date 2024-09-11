// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"math"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	diskUsageSensorIcon  = "mdi:harddisk"
	diskUsageSensorUnits = "%"
)

type diskUsageSensor struct {
	mount *mount
	linux.Sensor
}

func (d *diskUsageSensor) Name() string {
	return "Mountpoint " + d.mount.mountpoint + " Usage"
}

func (d *diskUsageSensor) ID() string {
	if d.mount.mountpoint == "/" {
		return "mountpoint_root"
	}

	return "mountpoint" + strings.ReplaceAll(d.mount.mountpoint, "/", "_")
}

func (d *diskUsageSensor) Attributes() map[string]any {
	return d.mount.attributes
}

//nolint:mnd
func newDiskUsageSensor(mount *mount) *diskUsageSensor {
	mount.attributes["data_source"] = linux.DataSrcProcfs

	usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:forcetypeassert
	mount.attributes["blocks_used"] = usedBlocks
	usedPc := float64(usedBlocks) / float64(mount.attributes[mountAttrBlocksTotal].(uint64)) * 100 //nolint:forcetypeassert

	return &diskUsageSensor{
		Sensor: linux.Sensor{
			IconString:      diskUsageSensorIcon,
			StateClassValue: types.StateClassTotal,
			UnitsString:     diskUsageSensorUnits,
			Value:           math.Round(float64(usedPc)/0.05) * 0.05,
		},
		mount: mount,
	}
}
