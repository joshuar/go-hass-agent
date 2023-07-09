// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package location

import (
	"context"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/api"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

// LocationUpdate represents a location update from a platform/device. It
// provides a bridge between the platform/device specific location info and Home
// Assistant.
type Update interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

// MarshalUpdate will take a LocationUpdate and marshal it into a
// hass.LocationUpdate that can be sent as a request to HA
func MarshalUpdate(l Update) *hass.LocationUpdate {
	return &hass.LocationUpdate{
		Gps:              l.Gps(),
		GpsAccuracy:      l.GpsAccuracy(),
		Battery:          l.Battery(),
		Speed:            l.Speed(),
		Altitude:         l.Altitude(),
		Course:           l.Course(),
		VerticalAccuracy: l.VerticalAccuracy(),
	}
}

func SendUpdate(ctx context.Context, l Update) {
	respCh := make(chan api.Response, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		api.ExecuteRequest(ctx, MarshalUpdate(l), respCh)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(respCh)
		response := <-respCh
		if response.Error() != nil {
			log.Debug().Err(response.Error()).
				Msg("Failed to update location.")
		} else {
			log.Debug().Msg("Location Updated.")
		}
	}()
}
