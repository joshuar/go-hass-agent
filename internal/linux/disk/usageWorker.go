// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package disk

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID = "disk_usage_sensors"
)

type usageWorker struct{}

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
	return basePreferencesID + "." + usageWorkerPreferencesID
}

func (w *usageWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		UpdateInterval: usageUpdateInterval.String(),
	}
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	usageWorker := &usageWorker{}

	prefs, err := preferences.LoadWorker(ctx, usageWorker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
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
