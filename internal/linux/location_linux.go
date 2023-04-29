// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/maltegrosse/go-geoclue2"
	"github.com/rs/zerolog/log"
)

const (
	appID = "org.joshuar.go-hass-agent"
)

type linuxLocation struct {
	latitude  float64
	longitude float64
	accuracy  float64
	speed     float64
	altitude  float64
}

// linuxLocation implements hass.LocationUpdate

func (l *linuxLocation) Gps() []float64 {
	return []float64{l.latitude, l.longitude}
}

func (l *linuxLocation) GpsAccuracy() int {
	return int(l.accuracy)
}

func (l *linuxLocation) Battery() int {
	return 0
}

func (l *linuxLocation) Speed() int {
	return int(l.speed)
}

func (l *linuxLocation) Altitude() int {
	return int(l.altitude)
}

func (l *linuxLocation) Course() int {
	return 0
}

func (l *linuxLocation) VerticalAccuracy() int {
	return 0
}

func LocationUpdater(ctx context.Context, locationInfoCh chan interface{}) {
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return
	}

	if deviceAPI.dBusSystem == nil {
		log.Debug().Caller().
			Msg("No system bus connection. Location sensor unavailable.")
		return
	}

	locationInfo := &linuxLocation{}

	gcm, err := geoclue2.NewGeoclueManager()
	if err != nil {
		log.Debug().Err(err).
			Msg("Could not create geoclue interface.")
	}

	client, err := gcm.GetClient()
	if err != nil {
		log.Debug().Err(err).
			Msg("Could not create geoclue interface.")
	}

	if err := client.SetDesktopId(appID); err != nil {
		log.Debug().Err(err).
			Msg("Could not create geoclue interface.")
	}

	if err := client.Start(); err != nil {
		log.Debug().Err(err).
			Msg("Could not create geoclue interface.")
	}

	c := client.SubscribeLocationUpdated()
	for {
		select {
		case v := <-c:
			log.Debug().Caller().Msg("Location update received.")
			_, location, _ := client.ParseLocationUpdated(v)
			locationInfo.latitude, _ = location.GetLatitude()
			locationInfo.longitude, _ = location.GetLongitude()
			locationInfo.accuracy, _ = location.GetAccuracy()
			locationInfo.speed, _ = location.GetSpeed()
			locationInfo.altitude, _ = location.GetAltitude()
			locationInfoCh <- locationInfo
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Stopping Linux location updater.")
			gcm.DeleteClient(client)
			return
		}
	}
}
