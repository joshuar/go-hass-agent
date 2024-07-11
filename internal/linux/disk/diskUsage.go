// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package disk

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/disk"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 10 * time.Second

	usageWorkerID = "disk_usage_sensors"
)

type diskUsageSensor struct {
	stats *disk.UsageStat
	linux.Sensor
}

//nolint:mnd
func newDiskUsageSensor(stat *disk.UsageStat) *diskUsageSensor {
	return &diskUsageSensor{
		Sensor: linux.Sensor{
			IconString:      "mdi:harddisk",
			StateClassValue: types.StateClassTotal,
			UnitsString:     "%",
			Value:           math.Round(stat.UsedPercent/0.05) * 0.05,
		},
		stats: stat,
	}
}

func (d *diskUsageSensor) Name() string {
	return "Mountpoint " + d.stats.Path + " Usage"
}

func (d *diskUsageSensor) ID() string {
	if d.stats.Path == "/" {
		return "mountpoint_root"
	}

	return "mountpoint" + strings.ReplaceAll(d.stats.Path, "/", "_")
}

func (d *diskUsageSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["data_source"] = linux.DataSrcProcfs
	attributes["stats"] = d.stats

	return attributes
}

type usageWorker struct {
	logger *slog.Logger
}

func (w *usageWorker) Interval() time.Duration { return usageUpdateInterval }

func (w *usageWorker) Jitter() time.Duration { return usageUpdateJitter }

func (w *usageWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of physical partitions: %w", err)
	}

	sensors := make([]sensor.Details, 0, len(partitions))

	for _, partition := range partitions {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			w.logger.Warn("Failed to get usage info for mountpount", "mountpoint", partition.Mountpoint, "error", err.Error())

			continue
		}

		sensors = append(sensors, newDiskUsageSensor(usage))
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
