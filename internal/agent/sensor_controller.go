// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package agent

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// runSensorWorkers will start all the sensor worker functions for all sensor
// controllers passed in. It returns a single merged channel of sensor updates.
//
//nolint:gocognit
func runSensorWorkers(ctx context.Context, controllers ...SensorController) {
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

	prefs, err := preferences.Load(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Could not start sensor controller.", slog.Any("error", err))
		return
	}

	hassclient, err := hass.NewClient(ctx)
	if err != nil {
		logging.FromContext(ctx).Debug("Cannot create Home Assistant client.", slog.Any("error", err))
		return
	}

	hassclient.Endpoint(prefs.RestAPIURL(), hass.DefaultTimeout)

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
					if err := hassclient.ProcessSensor(ctx, details); err != nil {
						logging.FromContext(ctx).Error("Process sensor failed.", slog.Any("error", err))
					}
				}(details)
			}

			return
		}
	}
}
