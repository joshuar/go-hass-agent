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

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID = "disk_usage_sensors"
)

type usageWorker struct {
	logger *slog.Logger
}

func (w *usageWorker) Interval() time.Duration { return usageUpdateInterval }

func (w *usageWorker) Jitter() time.Duration { return usageUpdateJitter }

func (w *usageWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	mounts, err := getMounts()
	if err != nil {
		return nil, fmt.Errorf("could not get mount points: %w", err)
	}

	sensors := make([]sensor.Details, 0, len(mounts))

	for _, mount := range mounts {
		sensors = append(sensors, newDiskUsageSensor(mount))
	}

	return sensors, nil
}

func NewUsageWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &usageWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", usageWorkerID)),
			},
			WorkerID: usageWorkerID,
		},
		nil
}
