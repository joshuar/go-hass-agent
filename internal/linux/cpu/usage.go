// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type cpuUsageSensor struct {
	linux.Sensor
}

type usageWorker struct{}

func (w *usageWorker) Interval() time.Duration { return 10 * time.Second }

func (w *usageWorker) Jitter() time.Duration { return time.Second }

func (w *usageWorker) Sensors(ctx context.Context, d time.Duration) ([]sensor.Details, error) {
	usage, err := cpu.PercentWithContext(ctx, d, false)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve CPU usage: %w", err)
	}
	s := &cpuUsageSensor{}
	s.IconString = "mdi:chip"
	s.UnitsString = "%"
	s.SensorSrc = linux.DataSrcProcfs
	s.StateClassValue = types.StateClassMeasurement
	s.Value = usage[0]
	s.SensorTypeValue = linux.SensorCPUPc
	return []sensor.Details{s}, nil
}

func NewUsageWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "CPU Usage Sensor",
			WorkerDesc: "System CPU usage as a percentage.",
			Value:      &usageWorker{},
		},
		nil
}
