// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type HassClient interface {
	ProcessSensor(ctx context.Context, details sensor.Details) error
	SensorList() []string
	GetSensor(id string) (sensor.Details, error)
	HassVersion(ctx context.Context) string
	Endpoint(url string, timeout time.Duration)
}

// runSensorWorkers will start all the sensor worker functions for all sensor
// controllers passed in. It returns a single merged channel of sensor updates.
//
//nolint:gocognit
func (agent *Agent) runSensorWorkers(ctx context.Context, controllers ...SensorController) {
	var sensorCh []<-chan sensor.Details

	for _, controller := range controllers {
		ch, err := controller.StartAll(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Start controller had errors.", slog.Any("errors", err))
		} else {
			sensorCh = append(sensorCh, ch)
		}
	}

	if len(sensorCh) == 0 {
		logging.FromContext(ctx).Warn("No workers were started by any controllers.")
		return
	}

	logging.FromContext(ctx).Debug("Processing sensor updates.")

	for {
		select {
		case <-ctx.Done():
			logging.FromContext(ctx).Debug("Stopping all sensor controllers.")

			for _, controller := range controllers {
				if err := controller.StopAll(); err != nil {
					logging.FromContext(ctx).Warn("Stop controller had errors.", slog.Any("error", err))
				}
			}

			return
		default:
			for details := range mergeCh(ctx, sensorCh...) {
				go func(details sensor.Details) {
					if err := agent.hass.ProcessSensor(ctx, details); err != nil {
						logging.FromContext(ctx).Error("Process sensor failed.", slog.Any("error", err))
					}
				}(details)
			}

			return
		}
	}
}
