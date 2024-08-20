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
//nolint:cyclop
func (agent *Agent) runSensorWorkers(ctx context.Context, controllers ...SensorController) {
	var sensorCh []<-chan sensor.Details

	for _, controller := range controllers {
		ch, err := controller.StartAll(ctx)
		if err != nil {
			agent.logger.Warn("Start controller had errors.", "errors", err.Error())
		} else {
			sensorCh = append(sensorCh, ch)
		}
	}

	if len(sensorCh) == 0 {
		agent.logger.Warn("No workers were started by any controllers.")

		return
	}

	agent.logger.Debug("Processing sensor updates.")

	for {
		select {
		case <-ctx.Done():
			agent.logger.Debug("Stopping all sensor controllers.")

			for _, controller := range controllers {
				if err := controller.StopAll(); err != nil {
					agent.logger.Warn("Stop controller had errors.", "error", err.Error())
				}
			}

			return
		default:
			for details := range mergeCh(ctx, sensorCh...) {
				go func(details sensor.Details) {
					if err := agent.hass.ProcessSensor(ctx, details); err != nil {
						agent.logger.Error("Process sensor failed.", slog.Any("error", err))
					}
				}(details)
			}

			return
		}
	}
}
