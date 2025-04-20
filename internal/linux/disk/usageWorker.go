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
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID   = "disk_usage_sensors"
	usageWorkerDesc = "Disk usage stats"
)

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
		usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,errcheck,forcetypeassert
		usedPc := float64(usedBlocks) / float64(mount.attributes[mountAttrBlocksTotal].(uint64)) * 100                 //nolint:errcheck,forcetypeassert

		if math.IsNaN(usedPc) {
			continue
		}

		diskUsageSensor, err := newDiskUsageSensor(ctx, mount, usedPc)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not generate usage sensor.", slog.Any("error", err))
			continue
		}
		w.OutCh <- *diskUsageSensor
	}
	return nil
}

func (w *usageWorker) PreferencesID() string {
	return usageWorkerPreferencesID
}

func (w *usageWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		UpdateInterval: usageUpdateInterval.String(),
	}
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

	prefs, err := preferences.LoadWorker(usageWorker)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
	}
	usageWorker.prefs = prefs

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", usageWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", usageUpdateInterval.String()))

		pollInterval = usageUpdateInterval
	}
	usageWorker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, usageUpdateJitter)

	return usageWorker, nil
}
