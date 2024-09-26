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
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
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
			sensor.Entity{
				Name:     "Distribution Name",
				Category: types.CategoryDiagnostic,
				EntityState: &sensor.EntityState{
					ID:    "distribution_name",
					State: distro,
					Icon:  "mdi:linux",
					Attributes: map[string]any{
						"data_source": linux.DataSrcProcfs,
					},
				},
			},
			sensor.Entity{
				Name:     "Distribution Version",
				Category: types.CategoryDiagnostic,
				EntityState: &sensor.EntityState{
					ID:    "distribution_version",
					State: version,
					Icon:  "mdi:numeric",
					Attributes: map[string]any{
						"data_source": linux.DataSrcProcfs,
					},
				},
			},
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
			sensor.Entity{
				Name:     "Kernel Version",
				Category: types.CategoryDiagnostic,
				EntityState: &sensor.EntityState{
					ID:    "kernel_version",
					State: kernelVersion,
					Icon:  "mdi:chip",
					Attributes: map[string]any{
						"data_source": linux.DataSrcProcfs,
					},
				},
			},
		)
	}

	return sensors, nil
}

func NewInfoWorker(_ context.Context) (*linux.OneShotSensorWorker, error) {
	worker := linux.NewOneShotWorker(infoWorkerID)
	worker.OneShotType = &infoWorker{}

	return worker, nil
}
