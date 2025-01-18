// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"math"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	diskUsageSensorIcon  = "mdi:harddisk"
	diskUsageSensorUnits = "%"
)

//nolint:mnd
func newDiskUsageSensor(mount *mount) sensor.Entity {
	mount.attributes["data_source"] = linux.DataSrcProcfs

	usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,errcheck,forcetypeassert
	mount.attributes["blocks_used"] = usedBlocks
	usedPc := float64(usedBlocks) / float64(mount.attributes[mountAttrBlocksTotal].(uint64)) * 100 //nolint:errcheck,forcetypeassert

	var id string

	if mount.mountpoint == "/" {
		id = "mountpoint_root"
	} else {
		id = "mountpoint" + strings.ReplaceAll(mount.mountpoint, "/", "_")
	}

	return sensor.NewSensor(
		sensor.WithName("Mountpoint "+mount.mountpoint+" Usage"),
		sensor.WithID(id),
		sensor.WithUnits(diskUsageSensorUnits),
		sensor.WithStateClass(types.StateClassTotal),
		sensor.WithState(
			sensor.WithIcon(diskUsageSensorIcon),
			sensor.WithValue(math.Round(float64(usedPc)/0.05)*0.05),
			sensor.WithAttributes(mount.attributes),
		),
	)
}
