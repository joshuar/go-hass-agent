// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	usageUpdateInterval = 10 * time.Second
	usageUpdateJitter   = time.Second

	usageWorkerID = "cpu_usage_sensors"
)

type cpuUsageSensor struct {
	linux.Sensor
}

type usageWorker struct{}

func (w *usageWorker) Interval() time.Duration { return usageUpdateInterval }

func (w *usageWorker) Jitter() time.Duration { return usageUpdateJitter }

func (w *usageWorker) Sensors(ctx context.Context, d time.Duration) ([]sensor.Details, error) {
	usage, err := cpu.PercentWithContext(ctx, d, false)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve CPU usage: %w", err)
	}

	newSensor := &cpuUsageSensor{}
	newSensor.IconString = "mdi:chip"
	newSensor.UnitsString = "%"
	newSensor.SensorSrc = linux.DataSrcProcfs
	newSensor.StateClassValue = types.StateClassMeasurement
	newSensor.Value = usage[0]
	newSensor.SensorTypeValue = linux.SensorCPUPc

	return []sensor.Details{newSensor}, nil
}

func NewUsageWorker(_ context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value:    &usageWorker{},
			WorkerID: usageWorkerID,
		},
		nil
}
