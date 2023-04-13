// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/maltegrosse/go-geoclue2"
	"github.com/rs/zerolog/log"
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

func LocationUpdater(ctx context.Context, appID string, locationInfoCh chan interface{}) {

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor network.")
		return
	}

	if deviceAPI.dBusSystem == nil {
		log.Debug().
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
			_, location, err := client.ParseLocationUpdated(v)
			logging.CheckError(err)

			locationInfo.latitude, err = location.GetLatitude()
			logging.CheckError(err)

			locationInfo.longitude, err = location.GetLongitude()
			logging.CheckError(err)

			locationInfo.accuracy, err = location.GetAccuracy()
			logging.CheckError(err)

			locationInfo.speed, err = location.GetSpeed()
			logging.CheckError(err)

			locationInfo.altitude, err = location.GetAltitude()
			logging.CheckError(err)

			locationInfoCh <- locationInfo
		case <-ctx.Done():
			log.Debug().Caller().
				Msg("Stopping Linux location updater.")
			gcm.DeleteClient(client)
			return
		}
	}
}
