// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package system

import (
	"bufio"
	"context"
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
	uptimePollInterval = 15 * time.Minute
	uptimePollJitter   = time.Minute

	uptimeWorkerID = "time"
)

type uptimeWorker struct{}

func (w *uptimeWorker) UpdateDelta(_ time.Duration) {}

func (w *uptimeWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	return []sensor.Entity{
			sensor.NewSensor(
				sensor.WithName("Uptime"),
				sensor.AsDiagnostic(),
				sensor.WithDeviceClass(types.SensorDeviceClassDuration),
				sensor.WithStateClass(types.StateClassMeasurement),
				sensor.WithUnits("h"),
				sensor.WithState(
					sensor.WithID("uptime"),
					sensor.WithIcon("mdi:restart"),
					sensor.WithValue(w.getUptime(ctx)/60/60),
					sensor.WithDataSourceAttribute(linux.ProcFSRoot),
					sensor.WithAttribute("native_unit_of_measurement", "h"),
				),
			),
		},
		nil
}

// getUptime retrieve the uptime of the device running Go Hass Agent, in
// seconds. If the uptime cannot be retrieved, it will return 0.
func (w *uptimeWorker) getUptime(ctx context.Context) float64 {
	data, err := os.Open(linux.UptimeFile)
	if err != nil {
		logging.FromContext(ctx).Debug("Unable to retrieve uptime.", slog.Any("error", err))

		return 0
	}

	defer data.Close()

	line := bufio.NewScanner(data)
	line.Split(bufio.ScanWords)

	if !line.Scan() {
		logging.FromContext(ctx).Debug("Could not parse uptime.")

		return 0
	}

	uptimeValue, err := strconv.ParseFloat(line.Text(), 64)
	if err != nil {
		logging.FromContext(ctx).Debug("Could not parse uptime.")

		return 0
	}

	return uptimeValue
}

func NewUptimeTimeWorker(_ context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingSensorWorker(uptimeWorkerID, uptimePollInterval, uptimePollJitter)
	worker.PollingSensorType = &uptimeWorker{}

	return worker, nil
}
