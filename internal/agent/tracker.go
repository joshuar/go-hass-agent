// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/sensors"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runSensorTracker(ctx context.Context) {
	log.Debug().Caller().Msg("Starting sensor tracker.")
	registryPath, err := agent.extraStoragePath("sensorRegistry")
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to store registry on disk, trying in-memory store.")
	}

	tracker = sensors.NewSensorTracker(ctx, registryPath)
	if tracker == nil {
		log.Debug().Msg("Unable to create a sensor tracker.")
		return
	}
	updateCh := make(chan interface{})
	go tracker.StartWorkers(ctx, updateCh)
	// Sensors are tracked in a map to handle registration and
	// disabling/enabling. Updates are sent to Home Assistant.
	for {
		select {
		case data := <-updateCh:
			switch data := data.(type) {
			case hass.SensorUpdate:
				go tracker.Update(ctx, data)
			case hass.LocationUpdate:
				l := hass.MarshalLocationUpdate(data)
				go hass.APIRequest(ctx, l)
			default:
				log.Debug().Caller().
					Msgf("Got unexpected status update %v", data)
			}
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Stopping sensor tracking.")
			return
		}
	}
}
