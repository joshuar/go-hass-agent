// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/location"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runSensorTracker(ctx context.Context) {
	registryPath, err := agent.extraStoragePath("sensorRegistry")
	if err != nil {
		log.Warn().Err(err).
			Msg("Unable to store registry on disk, will attempt in-memory store.")
	}

	sensorTracker = tracker.NewSensorTracker(ctx, registryPath.Path())
	if sensorTracker == nil {
		log.Error().Msg("Unable to create a sensor tracker.")
		return
	}
	updateCh := make(chan interface{})
	go sensorTracker.StartWorkers(ctx, updateCh)
	// Sensors are tracked in a map to handle registration and
	// disabling/enabling. Updates are sent to Home Assistant.
	for {
		select {
		case data := <-updateCh:
			switch data := data.(type) {
			case tracker.Sensor:
				go sensorTracker.Update(ctx, data)
			case location.Update:
				go location.SendUpdate(ctx, data)
			default:
				log.Warn().
					Msgf("Got unexpected status update %v", data)
			}
		case <-ctx.Done():
			log.Debug().
				Msg("Stopping sensor tracking.")
			return
		}
	}
}
