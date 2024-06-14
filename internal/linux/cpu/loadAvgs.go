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

	"github.com/shirou/gopsutil/v3/load"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	loadAvgIcon = "mdi:chip"
	loadAvgUnit = "load"

	loadAvgUpdateInterval = time.Minute
	loadAvgUpdateJitter   = 5 * time.Second
)

type loadavgSensor struct {
	linux.Sensor
}

type loadAvgsSensorWorker struct{}

func (w *loadAvgsSensorWorker) Interval() time.Duration { return loadAvgUpdateInterval }

func (w *loadAvgsSensorWorker) Jitter() time.Duration { return loadAvgUpdateJitter }

//nolint:exhaustive,exhaustruct,mnd
func (w *loadAvgsSensorWorker) Sensors(ctx context.Context, _ time.Duration) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, 3)

	loadAvgs, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("problem fetching load averages: %w", err)
	}

	for _, loadType := range []linux.SensorTypeValue{linux.SensorLoad1, linux.SensorLoad5, linux.SensorLoad15} {
		newSensor := &loadavgSensor{}
		newSensor.IconString = loadAvgIcon
		newSensor.UnitsString = loadAvgUnit
		newSensor.SensorSrc = linux.DataSrcProcfs
		newSensor.StateClassValue = types.StateClassMeasurement

		switch loadType {
		case linux.SensorLoad1:
			newSensor.Value = loadAvgs.Load1
			newSensor.SensorTypeValue = linux.SensorLoad1
		case linux.SensorLoad5:
			newSensor.Value = loadAvgs.Load5
			newSensor.SensorTypeValue = linux.SensorLoad5
		case linux.SensorLoad15:
			newSensor.Value = loadAvgs.Load15
			newSensor.SensorTypeValue = linux.SensorLoad15
		}

		sensors = append(sensors, newSensor)
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
