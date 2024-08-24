// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	uptimeInterval = 15 * time.Minute
	uptimeJitter   = time.Minute

	timeWorkerID = "time_sensors"
)

type timeSensor struct {
	linux.Sensor
}

type timeWorker struct {
	logger       *slog.Logger
	boottime     time.Time
	boottimeSent bool
}

func (w *timeWorker) Interval() time.Duration { return uptimeInterval }

func (w *timeWorker) Jitter() time.Duration { return uptimeJitter }

func (w *timeWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	var sensors []sensor.Details

	// Send the uptime.
	sensors = append(sensors, &timeSensor{
		linux.Sensor{
			SensorTypeValue:  linux.SensorUptime,
			Value:            w.getUptime() / 60 / 60, //nolint:mnd
			IsDiagnostic:     true,
			UnitsString:      "h",
			IconString:       "mdi:restart",
			DeviceClassValue: types.DeviceClassDuration,
			StateClassValue:  types.StateClassMeasurement,
		},
	})

	// Send the boottime if we haven't already.
	if !w.boottimeSent {
		sensors = append(sensors, &timeSensor{
			linux.Sensor{
				SensorTypeValue:  linux.SensorBoottime,
				Value:            w.boottime,
				IsDiagnostic:     true,
				IconString:       "mdi:restart",
				DeviceClassValue: types.DeviceClassDate,
			},
		})
		w.boottimeSent = true
	}

	return sensors, nil
}

// getUptime retrieve the uptime of the device running Go Hass Agent, in
// seconds. If the uptime cannot be retrieved, it will return 0.
func (w *timeWorker) getUptime() float64 {
	data, err := os.Open(linux.UptimeFile)
	if err != nil {
		w.logger.Debug("Unable to retrieve uptime.", slog.Any("error", err))

		return 0
	}

	defer data.Close()

	line := bufio.NewScanner(data)
	line.Split(bufio.ScanWords)

	if !line.Scan() {
		w.logger.Debug("Could not parse uptime.")

		return 0
	}

	uptimeValue, err := strconv.ParseFloat(line.Text(), 64)
	if err != nil {
		w.logger.Debug("Could not parse uptime.")

		return 0
	}

	return uptimeValue
}

func NewTimeWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	return &linux.SensorWorker{
			Value: &timeWorker{
				logger:   logging.FromContext(ctx).WithGroup(timeWorkerID),
				boottime: boottime,
			},
			WorkerID: timeWorkerID,
		},
		nil
}
