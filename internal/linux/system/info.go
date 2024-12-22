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
)

const (
	infoWorkerID = "system_info"
)

type infoWorker struct{}

func (w *infoWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	var sensors []sensor.Entity

	// Get distribution name and version.
	distro, version, err := device.GetOSDetails()
	if err != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", infoWorkerID)).
			Warn("Could not retrieve distro details.", slog.Any("error", err))
	} else {
		sensors = append(sensors,
			sensor.NewSensor(
				sensor.WithName("Distribution Name"),
				sensor.WithID("distribution_name"),
				sensor.AsDiagnostic(),
				sensor.WithState(
					sensor.WithIcon("mdi:linux"),
					sensor.WithValue(distro),
					sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
				),
			),
			sensor.NewSensor(
				sensor.WithName("Distribution Version"),
				sensor.WithID("distribution_version"),
				sensor.AsDiagnostic(),
				sensor.WithState(
					sensor.WithIcon("mdi:numeric"),
					sensor.WithValue(version),
					sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
				),
			),
		)
	}

	// Get kernel version.
	kernelVersion, err := device.GetKernelVersion()
	if err != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", infoWorkerID)).
			Warn("Could not retrieve kernel version.", slog.Any("error", err))
	} else {
		sensors = append(sensors,
			sensor.NewSensor(
				sensor.WithName("Kernel Version"),
				sensor.WithID("kernel_version"),
				sensor.AsDiagnostic(),
				sensor.WithState(
					sensor.WithIcon("mdi:chip"),
					sensor.WithValue(kernelVersion),
					sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
				),
			),
		)
	}

	return sensors, nil
}

func NewInfoWorker(_ context.Context) (*linux.OneShotSensorWorker, error) {
	worker := linux.NewOneShotSensorWorker(infoWorkerID)
	worker.OneShotSensorType = &infoWorker{}

	return worker, nil
}
