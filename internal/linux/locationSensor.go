// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
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

func LocationUpdater(ctx context.Context, tracker device.SensorTracker) {
	locationUpdateHandler := func(s *dbus.Signal) {
		if s.Name == locationUpdatedSignal {
			if locationPath, ok := s.Body[1].(dbus.ObjectPath); ok {
				if err := tracker.UpdateSensors(ctx, newLocation(ctx, locationPath)); err != nil {
					log.Error().Err(err).Msg("Could not update location.")
				}
			}
		}
	}

	clientPath := NewBusRequest(ctx, SystemBus).
		Path(geocluePath).
		Destination(geoclueInterface).GetData(getClientCall).AsObjectPath()
	if !clientPath.IsValid() {
		log.Error().Msg("Could not set up a geoclue client.")
		return
	}
	locationRequest := NewBusRequest(ctx, SystemBus).Path(clientPath).Destination(geoclueInterface)

	if err := locationRequest.SetProp(desktopIDProp, dbus.MakeVariant(appID)); err != nil {
		log.Error().Err(err).Msg("Could not set a geoclue client id.")
		return
	}

	if err := locationRequest.SetProp(distanceThresholdProp, dbus.MakeVariant(uint32(0))); err != nil {
		log.Warn().Err(err).Msg("Could not set distance threshold for geoclue requests.")
	}

	if err := locationRequest.SetProp(timeThresholdProp, dbus.MakeVariant(uint32(0))); err != nil {
		log.Warn().Err(err).Msg("Could not set time threshold for geoclue requests.")
	}

	if err := locationRequest.Call(startCall); err != nil {
		log.Warn().Err(err).Msg("Could not start geoclue client.")
		return
	}

	log.Debug().Msg("Tracking location with geoclue.")

	go func() {
		<-ctx.Done()
		err := locationRequest.Call(stopCall)
		if err != nil {
			log.Debug().Caller().Err(err).Msg("Failed to stop location updater.")
			return
		}
	}()

	err := NewBusRequest(ctx, SystemBus).
		Path(clientPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(clientPath),
			dbus.WithMatchInterface(clientInterface),
		}).
		Event(locationUpdatedSignal).
		Handler(locationUpdateHandler).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Could not watch for geoclue updates.")
	}
}

func newLocation(ctx context.Context, locationPath dbus.ObjectPath) *linuxLocation {
	getProp := func(prop string) float64 {
		value, err := NewBusRequest(ctx, SystemBus).
			Path(locationPath).
			Destination(geoclueInterface).
			GetProp("org.freedesktop.GeoClue2.Location." + prop)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msgf("Could not retrieve %s.", prop)
			return 0
		} else {
			return value.Value().(float64)
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
