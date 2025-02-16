// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package disk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID = "disk_usage_sensors"
)

var ErrInitUsageWorker = errors.New("could not init usage worker")

type usageWorker struct{}

func (w *usageWorker) UpdateDelta(_ time.Duration) {}

func (w *usageWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	mounts, err := getMounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get mount points: %w", err)
	}

	sensors := make([]models.Entity, 0, len(mounts))

	for _, mount := range mounts {
		usedBlocks := mount.attributes[mountAttrBlocksTotal].(uint64) - mount.attributes[mountAttrBlocksFree].(uint64) //nolint:lll,errcheck,forcetypeassert
		usedPc := float64(usedBlocks) / float64(mount.attributes[mountAttrBlocksTotal].(uint64)) * 100                 //nolint:errcheck,forcetypeassert

		if math.IsNaN(usedPc) {
			continue
		}

		diskUsageSensor, err := newDiskUsageSensor(ctx, mount, usedPc)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not generate usage sensor.", slog.Any("error", err))
		} else {
			sensors = append(sensors, diskUsageSensor)
		}
	}

	return sensors, nil
}

func (w *usageWorker) PreferencesID() string {
	return usageWorkerPreferencesID
}

func (w *usageWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		UpdateInterval: usageUpdateInterval.String(),
	}
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	usageWorker := &usageWorker{}

	prefs, err := preferences.LoadWorker(usageWorker)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", usageWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", usageUpdateInterval.String()))

		pollInterval = usageUpdateInterval
	}

	worker := linux.NewPollingSensorWorker(usageWorkerID, pollInterval, usageUpdateJitter)
	worker.PollingSensorType = usageWorker

	return worker, nil
}
