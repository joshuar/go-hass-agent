// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	appID                  = "org.joshuar.go-hass-agent"
	geoclueInterface       = "org.freedesktop.GeoClue2"
	geoclueClientInterface = geoclueInterface + ".Client"
	geocluePath            = "/org/freedesktop/GeoClue2/Manager"
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
	clientPath := NewBusRequest(SystemBus).
		Path(geocluePath).
		Destination(geoclueInterface).
		GetData("org.freedesktop.GeoClue2.Manager.GetClient").AsObjectPath()
	if !clientPath.IsValid() {
		log.Debug().Caller().
			Msg("Could not set up geoclue client.")
		return
	}

	err := NewBusRequest(SystemBus).
		Path(clientPath).
		Destination(geoclueInterface).
		SetProp("org.freedesktop.GeoClue2.Client.DesktopId",
			dbus.MakeVariant(appID))
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not create geoclue client")
		return
	}

	err = NewBusRequest(SystemBus).
		Path(clientPath).
		Destination(geoclueInterface).
		SetProp("org.freedesktop.GeoClue2.Client.DistanceThreshold",
			dbus.MakeVariant(uint32(0)))
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not set distance threshold")
		return
	}

	err = NewBusRequest(SystemBus).
		Path(clientPath).
		Destination(geoclueInterface).
		SetProp("org.freedesktop.GeoClue2.Client.TimeThreshold",
			dbus.MakeVariant(uint32(0)))
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not set time threshold")
		return
	}

	locationUpdateHandler := func(s *dbus.Signal) {
		if s.Name == "org.freedesktop.GeoClue2.Client.LocationUpdated" {
			if locationPath, ok := s.Body[1].(dbus.ObjectPath); ok {
				locationInfoCh <- newLocation(locationPath)
			}
		}
	}

	err = NewBusRequest(SystemBus).
		Path(clientPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(clientPath),
			dbus.WithMatchInterface(geoclueClientInterface),
		}).
		Event("org.freedesktop.GeoClue2.Client.LocationUpdated").
		Handler(locationUpdateHandler).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create location DBus watch.")
		return
	}

	err = NewBusRequest(SystemBus).
		Path(clientPath).
		Destination(geoclueInterface).
		Call("org.freedesktop.GeoClue2.Client.Start")
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to start location updater.")
		return
	}

	go func() {
		<-ctx.Done()
		log.Debug().Caller().
			Msg("Stopping location updater.")
		err = NewBusRequest(SystemBus).
			Path(clientPath).
			Destination(geoclueInterface).
			Call("org.freedesktop.GeoClue2.Client.Stop")
		if err != nil {
			log.Debug().Caller().Err(err).
				Msg("Failed to stop location updater.")
			return
		}
	}()
}

func newLocation(locationPath dbus.ObjectPath) *linuxLocation {
	getProp := func(prop string) float64 {
		value, err := NewBusRequest(SystemBus).
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
