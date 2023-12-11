// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

const (
	appID            = "org.joshuar.go-hass-agent"
	geoclueInterface = "org.freedesktop.GeoClue2"
	clientInterface  = geoclueInterface + ".Client"
	geocluePath      = "/org/freedesktop/GeoClue2/Manager"

	startCall     = "org.freedesktop.GeoClue2.Client.Start"
	stopCall      = "org.freedesktop.GeoClue2.Client.Stop"
	getClientCall = "org.freedesktop.GeoClue2.Manager.GetClient"

	desktopIDProp         = "org.freedesktop.GeoClue2.Client.DesktopId"
	distanceThresholdProp = "org.freedesktop.GeoClue2.Client.DistanceThreshold"
	timeThresholdProp     = "org.freedesktop.GeoClue2.Client.TimeThreshold"

	locationUpdatedSignal = "org.freedesktop.GeoClue2.Client.LocationUpdated"
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

func LocationUpdater(ctx context.Context) chan tracker.Location {
	sensorCh := make(chan tracker.Location, 1)
	locationUpdateHandler := func(s *dbus.Signal) {
		if s.Name == locationUpdatedSignal {
			if locationPath, ok := s.Body[1].(dbus.ObjectPath); ok {
				sensorCh <- newLocation(ctx, locationPath)
			}
		}
	}

	clientPath := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(geocluePath).
		Destination(geoclueInterface).GetData(getClientCall).AsObjectPath()
	if !clientPath.IsValid() {
		log.Error().Msg("Could not set up a geoclue client.")
		close(sensorCh)
		return sensorCh
	}
	locationRequest := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).Path(clientPath).Destination(geoclueInterface)

	if err := locationRequest.SetProp(desktopIDProp, dbus.MakeVariant(appID)); err != nil {
		log.Error().Err(err).Msg("Could not set a geoclue client id.")
		close(sensorCh)
		return sensorCh
	}

	if err := locationRequest.SetProp(distanceThresholdProp, dbus.MakeVariant(uint32(0))); err != nil {
		log.Warn().Err(err).Msg("Could not set distance threshold for geoclue requests.")
	}

	if err := locationRequest.SetProp(timeThresholdProp, dbus.MakeVariant(uint32(0))); err != nil {
		log.Warn().Err(err).Msg("Could not set time threshold for geoclue requests.")
	}

	if err := locationRequest.Call(startCall); err != nil {
		log.Warn().Err(err).Msg("Could not start geoclue client.")
		close(sensorCh)
		return sensorCh
	}

	log.Debug().Msg("Tracking location with geoclue.")

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(clientPath),
			dbus.WithMatchInterface(clientInterface),
			dbus.WithMatchMember("LocationUpdated"),
		}).
		Handler(locationUpdateHandler).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Could not watch for geoclue updates.")
	}

	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		err := locationRequest.Call(stopCall)
		if err != nil {
			log.Debug().Caller().Err(err).Msg("Failed to stop location updater.")
			return
		}
	}()
	return sensorCh
}

func newLocation(ctx context.Context, locationPath dbus.ObjectPath) *linuxLocation {
	getProp := func(prop string) float64 {
		value, err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
			Path(locationPath).
			Destination(geoclueInterface).
			GetProp("org.freedesktop.GeoClue2.Location." + prop)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msgf("Could not retrieve %s.", prop)
			return 0
		} else {
			return dbushelpers.VariantToValue[float64](value)
		}
	}
	return &linuxLocation{
		latitude:  getProp("Latitude"),
		longitude: getProp("Longitude"),
		accuracy:  getProp("Accuracy"),
		speed:     getProp("Speed"),
		altitude:  getProp("Altitude"),
	}
}
