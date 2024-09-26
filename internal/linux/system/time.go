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

	timeWorkerID = "time"
)

type timeWorker struct {
	boottime     time.Time
	logger       *slog.Logger
	boottimeSent bool
}

func (w *timeWorker) UpdateDelta(_ time.Duration) {}

func (w *timeWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	var sensors []sensor.Entity

	// Send the uptime.
	sensors = append(sensors,
		sensor.Entity{
			Name:        "Uptime",
			Category:    types.CategoryDiagnostic,
			Units:       "h",
			DeviceClass: types.DeviceClassDuration,
			StateClass:  types.StateClassMeasurement,
			EntityState: &sensor.EntityState{
				ID:    "uptime",
				State: w.getUptime() / 60 / 60, //nolint:mnd
				Icon:  "mdi:restart",
				Attributes: map[string]any{
					"data_source":                linux.DataSrcProcfs,
					"native_unit_of_measurement": "h",
				},
			},
		})

	// Send the boottime if we haven't already.
	if !w.boottimeSent {
		sensors = append(sensors,
			sensor.Entity{
				Name:        "Last Reboot",
				Category:    types.CategoryDiagnostic,
				DeviceClass: types.DeviceClassTimestamp,
				EntityState: &sensor.EntityState{
					ID:    "last_reboot",
					State: w.boottime.Format(time.RFC3339),
					Icon:  "mdi:restart",
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
