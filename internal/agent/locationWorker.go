package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

type locationData interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

type location struct {
	data    locationData
	encrypt bool
}

func (l *location) RequestType() hass.RequestType {
	return hass.RequestTypeUpdateLocation
}

func (l *location) RequestData() interface{} {
	return struct {
		Gps              []float64 `json:"gps"`
		GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
		Battery          int       `json:"battery,omitempty"`
		Speed            int       `json:"speed,omitempty"`
		Altitude         int       `json:"altitude,omitempty"`
		Course           int       `json:"course,omitempty"`
		VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
	}{
		Gps:              l.data.Gps(),
		GpsAccuracy:      l.data.GpsAccuracy(),
		Battery:          l.data.Battery(),
		Speed:            l.data.Speed(),
		Altitude:         l.data.Altitude(),
		Course:           l.data.Course(),
		VerticalAccuracy: l.data.VerticalAccuracy(),
	}
}

func (l *location) IsEncrypted() bool {
	return l.encrypt
}

func (l *location) handleResponse(rawResponse interface{}) {
	if rawResponse == nil {
		log.Debug().Caller().Msg("No response data.")
	} else {
		log.Debug().Caller().Msgf("Location updated to %v", l.data.Gps())
	}
}

func (agent *Agent) runLocationWorker(ctx context.Context) {

	locationInfoCh := make(chan interface{})
	defer close(locationInfoCh)

	go device.LocationUpdater(agent.App.UniqueID(), locationInfoCh)

	log.Debug().Caller().Msg("Running location worker.")

	for {
		select {
		case loc := <-locationInfoCh:
			l := &location{
				data: loc.(locationData),
			}

			go hass.APIRequest(ctx, agent.config.APIURL, l, l.handleResponse)
		case <-ctx.Done():
			log.Debug().Caller().Msgf("Cleaning up location sensor.")
			return
		}
	}
}
