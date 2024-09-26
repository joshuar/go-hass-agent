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

func NewUsageWorker(_ context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(usageWorkerID, usageUpdateInterval, usageUpdateJitter)
	worker.PollingType = &usageWorker{}

	return worker, nil
}
