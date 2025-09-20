// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID   = "disk_usage_sensors"
	usageWorkerDesc = "Disk usage stats"
)

const (
	diskUsageSensorIcon  = "mdi:harddisk"
	diskUsageSensorUnits = "%"
)

var ErrNewDiskUsageSensor = errors.New("could not create disk usage sensor")

//nolint:mnd
func newDiskUsageSensor(ctx context.Context, mount *mount, value float64) models.Entity {
	mount.attributes["data_source"] = linux.DataSrcProcFS

	usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,forcetypeassert
	mount.attributes["blocks_used"] = usedBlocks

	var id string

	if mount.mountpoint == "/" {
		id = "mountpoint_root"
	} else {
		id = "mountpoint" + strings.ReplaceAll(mount.mountpoint, "/", "_")
	}

	return sensor.NewSensor(ctx,
		sensor.WithName("Mountpoint "+mount.mountpoint+" Usage"),
		sensor.WithID(id),
		sensor.WithUnits(diskUsageSensorUnits),
		sensor.WithStateClass(class.StateTotal),
		sensor.WithIcon(diskUsageSensorIcon),
		sensor.WithState(math.Round(value/0.05)*0.05),
		sensor.WithAttributes(mount.attributes),
	)
}

var ErrInitUsageWorker = errors.New("could not init usage worker")

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
)

type usageWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
	prefs *WorkerPrefs
}

func (w *usageWorker) Execute(ctx context.Context) error {
	mounts, err := getMounts(ctx)
	if err != nil {
		return fmt.Errorf("could not get mount points: %w", err)
	}

	for mount := range slices.Values(mounts) {
		usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,forcetypeassert
		usedPc := float64(usedBlocks) / float64(mount.attributes[mountAttrBlocksTotal].(uint64)) * 100                 //nolint:forcetypeassert

		if math.IsNaN(usedPc) {
			continue
		}
		w.OutCh <- newDiskUsageSensor(ctx, mount, usedPc)
	}
	return nil
}

func (w *usageWorker) IsDisabled() bool {
	return w.prefs.Disabled
}

func (w *usageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func NewUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	usageWorker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: usageUpdateInterval.String(),
	}
	var err error
	usageWorker.prefs, err = workers.LoadWorkerPreferences(usageWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
	}

	pollInterval, err := time.ParseDuration(usageWorker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", usageWorkerID),
			slog.String("given_interval", usageWorker.prefs.UpdateInterval),
			slog.String("default_interval", usageUpdateInterval.String()))

		pollInterval = usageUpdateInterval
	}
	usageWorker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, usageUpdateJitter)

	return usageWorker, nil
}
