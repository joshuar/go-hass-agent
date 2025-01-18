// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package disk

import (
	"context"
	"fmt"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID = "disk_usage_sensors"
)

type usageWorker struct {
	prefs *WorkerPrefs
}

func (w *usageWorker) UpdateDelta(_ time.Duration) {}

func (w *usageWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	mounts, err := getMounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get mount points: %w", err)
	}

	sensors := make([]sensor.Entity, 0, len(mounts))

	for _, mount := range mounts {
		sensors = append(sensors, newDiskUsageSensor(mount))
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
	var err error

	worker := linux.NewPollingSensorWorker(usageWorkerID, usageUpdateInterval, usageUpdateJitter)

	usageWorker := &usageWorker{}

	usageWorker.prefs, err = preferences.LoadWorker(ctx, usageWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if usageWorker.prefs.Disabled {
		return worker, nil
	}

	worker.PollingSensorType = usageWorker

	return worker, nil
}
