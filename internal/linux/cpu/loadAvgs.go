// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/load"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	loadAvgIcon = "mdi:chip"
	loadAvgUnit = "load"
)

type loadavgSensor struct {
	linux.Sensor
}

type loadAvgsSensorWorker struct{}

func (w *loadAvgsSensorWorker) Interval() time.Duration { return time.Minute }

func (w *loadAvgsSensorWorker) Jitter() time.Duration { return 5 * time.Second }

func (w *loadAvgsSensorWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	var sensors []sensor.Details

	latest, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("problem fetching load averages: %w", err)
	}

	for _, loadType := range []linux.SensorTypeValue{linux.SensorLoad1, linux.SensorLoad5, linux.SensorLoad15} {
		l := &loadavgSensor{}
		l.IconString = loadAvgIcon
		l.UnitsString = loadAvgUnit
		l.SensorSrc = linux.DataSrcProcfs
		l.StateClassValue = types.StateClassMeasurement
		switch loadType {
		case linux.SensorLoad1:
			l.Value = latest.Load1
			l.SensorTypeValue = linux.SensorLoad1
		case linux.SensorLoad5:
			l.Value = latest.Load5
			l.SensorTypeValue = linux.SensorLoad5
		case linux.SensorLoad15:
			l.Value = latest.Load15
			l.SensorTypeValue = linux.SensorLoad15
		}
		sensors = append(sensors, l)
	}
	return sensors, nil
}

func NewLoadAvgWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Load Average Sensors",
			WorkerDesc: "The canonical 1min, 5min and 15min load averages.",
			Value:      &loadAvgsSensorWorker{},
		},
		nil
}
