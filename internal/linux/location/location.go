// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package location

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	managerPath           = "/org/freedesktop/GeoClue2/Manager"
	geoclueInterface      = "org.freedesktop.GeoClue2"
	clientInterface       = geoclueInterface + ".Client"
	managerInterface      = geoclueInterface + ".Manager"
	locationInterface     = geoclueInterface + ".Location"
	startCall             = clientInterface + ".Start"
	stopCall              = clientInterface + ".Stop"
	getClientCall         = managerInterface + ".GetClient"
	desktopIDProp         = clientInterface + ".DesktopId"
	distanceThresholdProp = clientInterface + ".DistanceThreshold"
	timeThresholdProp     = clientInterface + ".TimeThreshold"
	locationUpdatedSignal = clientInterface + ".LocationUpdated"
)

func Updater(ctx context.Context) chan *hass.LocationData {
	sensorCh := make(chan *hass.LocationData)
	locationUpdateHandler := func(s *dbus.Signal) {
		if s.Name == locationUpdatedSignal {
			if locationPath, ok := s.Body[1].(dbus.ObjectPath); ok {
				go func() {
					sensorCh <- newLocation(ctx, locationPath)
				}()
			}
		}
	}

	clientReq := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(managerPath).
		Destination(geoclueInterface)

	var clientPath dbus.ObjectPath
	var err error

	clientPath, err = dbusx.GetData[dbus.ObjectPath](clientReq, getClientCall)
	if !clientPath.IsValid() || err != nil {
		log.Error().Err(err).Msg("Could not set up a geoclue client.")
		close(sensorCh)
		return sensorCh
	}
	locationRequest := dbusx.NewBusRequest(ctx, dbusx.SystemBus).Path(clientPath).Destination(geoclueInterface)

	if err = dbusx.SetProp(locationRequest, desktopIDProp, preferences.AppID); err != nil {
		log.Error().Err(err).Msg("Could not set a geoclue client id.")
		close(sensorCh)
		return sensorCh
	}

	if err = dbusx.SetProp(locationRequest, distanceThresholdProp, uint32(0)); err != nil {
		log.Warn().Err(err).Msg("Could not set distance threshold for geoclue requests.")
	}

	if err = dbusx.SetProp(locationRequest, timeThresholdProp, uint32(0)); err != nil {
		log.Warn().Err(err).Msg("Could not set time threshold for geoclue requests.")
	}

	if err = locationRequest.Call(startCall); err != nil {
		log.Warn().Err(err).Msg("Could not start geoclue client.")
		close(sensorCh)
		return sensorCh
	}

	log.Debug().Msg("Tracking location with geoclue.")

	err = dbusx.NewBusRequest(ctx, dbusx.SystemBus).
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

func newLocation(ctx context.Context, locationPath dbus.ObjectPath) *hass.LocationData {
	getProp := func(prop string) float64 {
		req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
			Path(locationPath).
			Destination(geoclueInterface)
		value, err := dbusx.GetProp[float64](req, locationInterface+"."+prop)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msgf("Could not retrieve %s.", prop)
			return 0
		}
		return value
	}
	return &hass.LocationData{
		Gps:         []float64{getProp("Latitude"), getProp("Longitude")},
		GpsAccuracy: int(getProp("Accuracy")),
		Speed:       int(getProp("Speed")),
		Altitude:    int(getProp("Altitude")),
	}
}
