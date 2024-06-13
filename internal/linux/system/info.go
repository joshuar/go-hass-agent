// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

type infoWorker struct{}

//nolint:exhaustruct
func (w *infoWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	// Get distribution name and version.
	distro, distroVersion := linux.GetDistroDetails()

	// Get kernel version.
	kernelVersion := linux.GetKernelVersion()

	return []sensor.Details{
			&linux.Sensor{
				SensorTypeValue: linux.SensorKernel,
				Value:           kernelVersion,
				IsDiagnostic:    true,
				IconString:      "mdi:chip",
				SensorSrc:       linux.DataSrcProcfs,
			},
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
		},
		nil
}

func NewInfoWorker(_ context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "System Info Sensors",
			WorkerDesc: "Sensors for kernel version, and Distribution name and version.",
			Value:      &infoWorker{},
		},
		nil
}
