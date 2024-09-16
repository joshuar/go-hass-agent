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
)

const (
	uptimeInterval = 15 * time.Minute
	uptimeJitter   = time.Minute

	timeWorkerID = "time_sensors"
)

type timeWorker struct {
	boottime     time.Time
	logger       *slog.Logger
	boottimeSent bool
}

func (w *timeWorker) UpdateDelta(_ time.Duration) {}

func (w *timeWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details

	// Send the uptime.
	sensors = append(sensors, &linux.Sensor{
		DisplayName:      "Uptime",
		UniqueID:         "uptime",
		Value:            w.getUptime() / 60 / 60, //nolint:mnd
		IsDiagnostic:     true,
		UnitsString:      "h",
		IconString:       "mdi:restart",
		DeviceClassValue: types.DeviceClassDuration,
		StateClassValue:  types.StateClassMeasurement,
	})

	// Send the boottime if we haven't already.
	if !w.boottimeSent {
		sensors = append(sensors, &linux.Sensor{
			DisplayName:      "Last Reboot",
			UniqueID:         "last_reboot",
			Value:            w.boottime.Format(time.RFC3339),
			IsDiagnostic:     true,
			IconString:       "mdi:restart",
			DeviceClassValue: types.DeviceClassTimestamp,
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

func NewTimeWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(timeWorkerID, uptimeInterval, uptimeJitter)

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	worker.PollingType = &timeWorker{
		boottime: boottime,
		logger:   logging.FromContext(ctx).With(slog.String("worker", timeWorkerID)),
	}

	return worker, nil
}
