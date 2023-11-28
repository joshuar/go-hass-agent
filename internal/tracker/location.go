// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"
)

// LocationUpdate represents a location update from a platform/device. It
// provides a bridge between the platform/device specific location info and Home
// Assistant.
type Location interface {
	Gps() []float64
	GpsAccuracy() int
	Battery() int
	Speed() int
	Altitude() int
	Course() int
	VerticalAccuracy() int
}

// marshalLocationUpdate will take a LocationUpdate and marshal it into a
// hass.LocationUpdate that can be sent as a request to HA
func marshalLocationUpdate(l Location) *hass.LocationUpdate {
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

func updateLocation(ctx context.Context, l Location) {
	response := <-api.ExecuteRequest(ctx, marshalLocationUpdate(l))
	switch r := response.(type) {
	case []byte:
		log.Debug().Msg("Location Updated.")
	case error:
		log.Warn().Err(r).Msg("Failed to update location.")
	default:
		log.Warn().Msgf("Unknown response type %T", r)
	}
}
