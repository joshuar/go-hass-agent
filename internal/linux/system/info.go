// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	infoWorkerID = "system_info_sensors"
)

type infoWorker struct {
	logger *slog.Logger
}

func (w *infoWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details

	// Get distribution name and version.
	distro, version, err := device.GetOSDetails()
	if err != nil {
		w.logger.Warn("Could not retrieve kernel version.", "error", err.Error())
	} else {
		sensors = append(sensors,
			&linux.Sensor{
				DisplayName:  "Distribution Name",
				Value:        distro,
				IsDiagnostic: true,
				IconString:   "mdi:linux",
				DataSource:   linux.DataSrcProcfs,
			},
			&linux.Sensor{
				DisplayName:  "Distribution Version",
				Value:        version,
				IsDiagnostic: true,
				IconString:   "mdi:numeric",
				DataSource:   linux.DataSrcProcfs,
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
				DisplayName:  "Kernel Version",
				Value:        kernelVersion,
				IsDiagnostic: true,
				IconString:   "mdi:chip",
				DataSource:   linux.DataSrcProcfs,
			},
		)
	}

	return sensors, nil
}

func NewInfoWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &infoWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", infoWorkerID)),
			},
			WorkerID: infoWorkerID,
		},
		nil
}
