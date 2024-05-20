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

	// The process to watch for location updates via D-Bus is tedious...

	var clientPath dbus.ObjectPath
	var err error
	clientPath, err = dbusx.GetData[dbus.ObjectPath](ctx, dbusx.SystemBus, managerPath, geoclueInterface, getClientCall)
	if !clientPath.IsValid() || err != nil {
		log.Error().Err(err).Msg("Could not set up a geoclue client.")
		close(sensorCh)
		return sensorCh
	}

	// Set our client ID.
	if err = dbusx.SetProp(ctx, dbusx.SystemBus, string(clientPath), geoclueInterface, desktopIDProp, preferences.AppID); err != nil {
		log.Error().Err(err).Msg("Could not set a geoclue client id.")
		close(sensorCh)
		return sensorCh
	}

	// Set a distance threshold.
	if err = dbusx.SetProp(ctx, dbusx.SystemBus, string(clientPath), geoclueInterface, distanceThresholdProp, uint32(0)); err != nil {
		log.Warn().Err(err).Msg("Could not set distance threshold for geoclue requests.")
	}

	// Set a time threshold.
	if err = dbusx.SetProp(ctx, dbusx.SystemBus, string(clientPath), geoclueInterface, timeThresholdProp, uint32(0)); err != nil {
		log.Warn().Err(err).Msg("Could not set time threshold for geoclue requests.")
	}

	// Request to start tracking location updates.
	if err = dbusx.Call(ctx, dbusx.SystemBus, string(clientPath), geoclueInterface, startCall); err != nil {
		log.Warn().Err(err).Msg("Could not start geoclue client.")
		close(sensorCh)
		return sensorCh
	}

	// Start our watch for the location update messages.
	eventCh, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Path:      string(clientPath),
		Interface: clientInterface,
		Names:     []string{"LocationUpdated"},
	})
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create location D-Bus watch.")
		close(sensorCh)
	}
	log.Debug().Msg("Tracking location with geoclue.")

	// Listen for the location updates and dispatch them as location sensor
	// updates.
	go func() {
		defer close(sensorCh)
		for {
			select {
			case <-ctx.Done():
				err := dbusx.Call(ctx, dbusx.SystemBus, string(clientPath), geoclueInterface, stopCall)
				if err != nil {
					log.Debug().Caller().Err(err).Msg("Failed to stop location updater.")
					return
				}
				return
			case event := <-eventCh:
				if locationPath, ok := event.Content[1].(dbus.ObjectPath); ok {
					go func() {
						sensorCh <- newLocation(ctx, locationPath)
					}()
				}
			}
		}
	}()
	return sensorCh
}

func newLocation(ctx context.Context, locationPath dbus.ObjectPath) *hass.LocationData {
	getProp := func(prop string) float64 {
		value, err := dbusx.GetProp[float64](ctx, dbusx.SystemBus, string(locationPath), geoclueInterface, locationInterface+"."+prop)
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
