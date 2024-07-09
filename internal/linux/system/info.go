// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package system

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type infoWorker struct{}

//nolint:exhaustruct
func (w *infoWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details

	// Get distribution name and version.
	distro, distroVersion, err := device.GetOSDetails()
	if err != nil {
		logging.FromContext(ctx).Warn("Could not retrieve distribution details.", "error", err.Error())
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
		logging.FromContext(ctx).Warn("Could not retrieve kernel version.", "error", err.Error())
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

func NewInfoWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "System Info Sensors",
			WorkerDesc: "Sensors for kernel version, and Distribution name and version.",
			Value:      &infoWorker{},
		},
		nil
}
