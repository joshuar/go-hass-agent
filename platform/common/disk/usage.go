// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package disk contains platform-independent logic for reporting disk
// mount usage as sensors. Each platform (linux, darwin, ...) supplies its
// own GetMountsFunc to enumerate mounts; everything else -- turning those
// mounts into sensors and running them as a polling worker -- is shared
// here.
package disk

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

const (
	SensorIcon  = "mdi:harddisk"
	SensorUnits = "%"
)

// Attribute keys stored on a Mount. Values that are read back by key
// elsewhere (here, or by a platform's own code) are constified; write-once
// values are not.
const (
	AttrDevice      = "device"
	AttrFs          = "filesystem_type"
	AttrOpts        = "mount_options"
	AttrBlockSize   = "block_size"
	AttrBlocksTotal = "blocks_total"
	AttrBlocksFree  = "blocks_free"
	AttrBlocksAvail = "blocks_available"
	AttrInodesTotal = "inodes_total"
	AttrInodesFree  = "inodes_free"
	AttrBytesTotal  = "bytes_total"
	AttrBytesFree   = "bytes_free"
)

// Mount represents a filesystem mount point and its usage attributes.
// Platform implementations populate Attributes with, at minimum,
// AttrBlockSize (as uint64), AttrBlocksTotal and AttrBlocksFree (both
// uint64), plus whatever other attributes are relevant on that platform.
type Mount struct {
	Attributes map[string]any
	Mountpoint string
}

func (m Mount) usedBlocks() uint64 {
	total, _ := m.Attributes[AttrBlocksTotal].(uint64) //nolint:forcetypeassert,errcheck
	free, _ := m.Attributes[AttrBlocksFree].(uint64)   //nolint:forcetypeassert,errcheck
	return total - free
}

// usedPercent returns the percentage of blocks used, or NaN if the total
// block count is unknown/zero (e.g. some pseudo filesystems).
func (m Mount) usedPercent() float64 {
	total, _ := m.Attributes[AttrBlocksTotal].(uint64) //nolint:forcetypeassert,errcheck
	if total == 0 {
		return math.NaN()
	}
	return float64(m.usedBlocks()) / float64(total) * 100
}

// newUsageSensor builds the disk usage sensor entity for a mount. dataSource
// identifies where the platform sourced its mount data from (e.g. "ProcFS",
// "getfsstat"), for the sensor's data_source attribute.
func newUsageSensor(ctx context.Context, m Mount, usedPc float64, dataSource string) models.Entity {
	blockSize, _ := m.Attributes[AttrBlockSize].(uint64) //nolint:forcetypeassert,errcheck
	usedBlocks := m.usedBlocks()

	m.Attributes["data_source"] = dataSource
	m.Attributes["blocks_used"] = usedBlocks
	m.Attributes["bytes_used"] = usedBlocks * blockSize

	var id string

	if m.Mountpoint == "/" {
		id = "mountpoint_root"
	} else {
		id = "mountpoint" + strings.ReplaceAll(m.Mountpoint, "/", "_")
	}

	return models.NewSensor(ctx,
		models.WithName("Mountpoint "+m.Mountpoint+" Usage"),
		models.WithID(id),
		models.WithUnits(SensorUnits),
		models.WithStateClass(models.StateTotal),
		models.WithIcon(SensorIcon),
		models.WithState(math.Round(usedPc/0.05)*0.05),
		models.WithAttributes(m.Attributes),
	)
}

// GetMountsFunc is implemented per-platform to enumerate the current
// filesystem mounts that should be reported on.
type GetMountsFunc func(ctx context.Context) ([]Mount, error)

var (
	_ quartz.Job                  = (*Worker)(nil)
	_ workers.PollingEntityWorker = (*Worker)(nil)
)

// Worker is a platform-independent polling entity worker that reports disk
// mount usage sensors. Platforms construct one with their own GetMounts
// implementation and DataSource label.
type Worker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	GetMounts  GetMountsFunc
	DataSource string
	Disabled   bool
}

func (w *Worker) IsDisabled() bool {
	return w.Disabled
}

func (w *Worker) Execute(ctx context.Context) error {
	mounts, err := w.GetMounts(ctx)
	if err != nil {
		return fmt.Errorf("could not get mount points: %w", err)
	}

	for _, m := range mounts {
		usedPc := m.usedPercent()
		if math.IsNaN(usedPc) {
			continue
		}
		w.OutCh <- newUsageSensor(ctx, m, usedPc, w.DataSource)
	}

	return nil
}

func (w *Worker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}
