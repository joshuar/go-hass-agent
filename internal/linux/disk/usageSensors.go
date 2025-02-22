// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	diskUsageSensorIcon  = "mdi:harddisk"
	diskUsageSensorUnits = "%"
)

var ErrNewDiskUsageSensor = errors.New("could not create disk usage sensor")

//nolint:mnd
func newDiskUsageSensor(ctx context.Context, mount *mount, value float64) (*models.Entity, error) {
	mount.attributes["data_source"] = linux.DataSrcProcfs

	usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,errcheck,forcetypeassert
	mount.attributes["blocks_used"] = usedBlocks

	var id string

	if mount.mountpoint == "/" {
		id = "mountpoint_root"
	} else {
		id = "mountpoint" + strings.ReplaceAll(mount.mountpoint, "/", "_")
	}

	usageSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Mountpoint "+mount.mountpoint+" Usage"),
		sensor.WithID(id),
		sensor.WithUnits(diskUsageSensorUnits),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(diskUsageSensorIcon),
		sensor.WithState(math.Round(value/0.05)*0.05),
		sensor.WithAttributes(mount.attributes),
	)
	if err != nil {
		return nil, errors.Join(ErrNewDiskUsageSensor, err)
	}

	return &usageSensor, nil
}
