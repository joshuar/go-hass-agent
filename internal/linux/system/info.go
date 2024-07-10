// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package system

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	infoWorkerID = "system_info_sensors"
)

type infoWorker struct {
	logger *slog.Logger
}

//nolint:exhaustruct
func (w *infoWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details

	// Get distribution name and version.
	distro, distroVersion, err := device.GetOSDetails()
	if err != nil {
		w.logger.Warn("Could not retrieve distribution details.", "error", err.Error())
	} else {
		sensors = append(sensors,
			&linux.Sensor{
				SensorTypeValue: linux.SensorDistribution,
				Value:           distro,
				IsDiagnostic:    true,
				IconString:      "mdi:linux",
				SensorSrc:       linux.DataSrcProcfs,
			},
			&linux.Sensor{
				SensorTypeValue: linux.SensorVersion,
				Value:           distroVersion,
				IsDiagnostic:    true,
				IconString:      "mdi:numeric",
				SensorSrc:       linux.DataSrcProcfs,
			},
		)
	}

	// Get kernel version.
	kernelVersion, err := device.GetKernelVersion()
	if err != nil {
		w.logger.Warn("Could not retrieve kernel version.", "error", err.Error())
	} else {
		sensors = append(sensors,
			&linux.Sensor{
				SensorTypeValue: linux.SensorKernel,
				Value:           kernelVersion,
				IsDiagnostic:    true,
				IconString:      "mdi:chip",
				SensorSrc:       linux.DataSrcProcfs,
			},
		)
	}

	return sensors, nil
}

func NewInfoWorker(ctx context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &infoWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", infoWorkerID)),
			},
			WorkerID: infoWorkerID,
		},
		nil
}
