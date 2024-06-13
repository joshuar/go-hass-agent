// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package location

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

type locationSensor struct {
	linux.Sensor
}

func (s *locationSensor) Name() string { return "Location" }

func (s *locationSensor) ID() string { return "location" }

type worker struct {
	clientPath dbus.ObjectPath
}

//nolint:exhaustruct
func (w *worker) Setup(ctx context.Context) *dbusx.Watch {
	var err error

	if err = dbusx.SetProp(ctx, dbusx.SystemBus, string(w.clientPath), geoclueInterface, desktopIDProp, preferences.AppID); err != nil {
		log.Error().Err(err).Msg("Could not set a geoclue client id.")

		return nil
	}

	// Set a distance threshold.
	if err = dbusx.SetProp(ctx, dbusx.SystemBus, string(w.clientPath), geoclueInterface, distanceThresholdProp, uint32(0)); err != nil {
		log.Warn().Err(err).Msg("Could not set distance threshold for geoclue requests.")
	}

	// Set a time threshold.
	if err = dbusx.SetProp(ctx, dbusx.SystemBus, string(w.clientPath), geoclueInterface, timeThresholdProp, uint32(0)); err != nil {
		log.Warn().Err(err).Msg("Could not set time threshold for geoclue requests.")
	}

	// Request to start tracking location updates.
	if err = dbusx.Call(ctx, dbusx.SystemBus, string(w.clientPath), geoclueInterface, startCall); err != nil {
		log.Warn().Err(err).Msg("Could not start geoclue client.")

		return nil
	}

	log.Debug().Msg("GeoClue client created.")

	return &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Path:      string(w.clientPath),
		Interface: clientInterface,
		Names:     []string{"LocationUpdated"},
	}
}

func (w *worker) Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				err := dbusx.Call(ctx, dbusx.SystemBus, string(w.clientPath), geoclueInterface, stopCall)
				if err != nil {
					log.Debug().Caller().Err(err).Msg("Failed to stop location updater.")

					return
				}

				return
			case event := <-triggerCh:
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

func (w *worker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

func NewLocationWorker(ctx context.Context) (*linux.SensorWorker, error) {
	// Don't run this worker if we are not running on a laptop.
	if linux.Chassis() != "laptop" {
		return nil, linux.ErrUnsupportedHardware
	}

	// Check if we can create a client, bail if we can't.
	clientPath, err := dbusx.GetData[dbus.ObjectPath](ctx, dbusx.SystemBus, managerPath, geoclueInterface, getClientCall)
	if !clientPath.IsValid() || err != nil {
		return nil, fmt.Errorf("could not set up a geoclue client: %w", err)
	}

	return &linux.SensorWorker{
			WorkerName: "Location Sensor",
			WorkerDesc: "Sensor for device location, from GeoClue.",
			Value: &worker{
				clientPath: clientPath,
			},
		},
		nil
}

//nolint:exhaustruct
func newLocation(ctx context.Context, locationPath dbus.ObjectPath) *locationSensor {
	getProp := func(prop string) float64 {
		value, err := dbusx.GetProp[float64](ctx, dbusx.SystemBus, string(locationPath), geoclueInterface, locationInterface+"."+prop)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msgf("Could not retrieve %s.", prop)

			return 0
		}

		return value
	}
	location := &locationSensor{}
	location.Value = &sensor.LocationRequest{
		Gps:         []float64{getProp("Latitude"), getProp("Longitude")},
		GpsAccuracy: int(getProp("Accuracy")),
		Speed:       int(getProp("Speed")),
		Altitude:    int(getProp("Altitude")),
	}

	return location
}
