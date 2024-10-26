// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"fmt"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	lastBootWorkerID = "boot_time_sensor"
)

type lastBootWorker struct {
	lastBoot time.Time
}

func (w *lastBootWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return []sensor.Entity{
			{
				Name:        "Last Reboot",
				Category:    types.CategoryDiagnostic,
				DeviceClass: types.SensorDeviceClassTimestamp,
				State: &sensor.State{
					ID:    "last_reboot",
					Value: w.lastBoot.Format(time.RFC3339),
					Icon:  "mdi:restart",
				},
			},
		},
		nil
}

func NewLastBootWorker(ctx context.Context) (*linux.OneShotSensorWorker, error) {
	worker := linux.NewOneShotSensorWorker(lastBootWorkerID)

	lastBoot, found := linux.CtxGetBoottime(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no lastBoot value", linux.ErrInvalidCtx)
	}

	worker.OneShotSensorType = &lastBootWorker{lastBoot: lastBoot}

	return worker, nil
}
