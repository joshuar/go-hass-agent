// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package location

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
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

	workerID = "location_sensor"
)

type locationWorker struct {
	getLocationProperty func(path, prop string) (float64, error)
	stopMethod          *dbusx.Method
	startMethod         *dbusx.Method
	triggerCh           chan dbusx.Trigger
}

//nolint:gocognit
func (w *locationWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	logger := logging.FromContext(ctx).With(slog.String("worker", workerID))

	err := w.startMethod.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start geoclue client: %w", err)
	}

	sensorCh := make(chan sensor.Entity)

	go func() {
		logger.Debug("Monitoring for location updates.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				if err := w.stopMethod.Call(ctx); err != nil {
					logger.Debug("Could not stop geoclue client.", slog.Any("error", err))
				}

				return
			case event := <-w.triggerCh:
				if locationPath, ok := event.Content[1].(string); ok {
					go func() {
						locationSensor, err := w.newLocation(locationPath)
						if err != nil {
							logger.Error("Could not update location.", slog.Any("error", err))
						} else {
							sensorCh <- locationSensor
						}
					}()
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *locationWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return nil, linux.ErrUnimplemented
}

func (w *locationWorker) newLocation(locationPath string) (sensor.Entity, error) {
	var warnings error

	latitude, err := w.getLocationProperty(locationPath, "Latitude")
	warnings = errors.Join(warnings, err)
	longitude, err := w.getLocationProperty(locationPath, "Longitude")
	warnings = errors.Join(warnings, err)
	speed, err := w.getLocationProperty(locationPath, "Speed")
	warnings = errors.Join(warnings, err)
	altitude, err := w.getLocationProperty(locationPath, "Altitude")
	warnings = errors.Join(warnings, err)
	accuracy, err := w.getLocationProperty(locationPath, "Accuracy")
	warnings = errors.Join(warnings, err)

	location := sensor.Entity{
		EntityState: &sensor.EntityState{
			State: &hass.LocationRequest{
				Gps:         []float64{latitude, longitude},
				GpsAccuracy: int(accuracy),
				Speed:       int(speed),
				Altitude:    int(altitude),
			},
		},
	}

	return location, warnings
}

func NewLocationWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(workerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	// Create a GeoClue client.
	clientPath, err := createClient(bus)
	if err != nil {
		return worker, fmt.Errorf("unable to create geoclue client: %w", err)
	}

	// Set threshold values.
	setThresholds(bus, clientPath)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(clientPath),
		dbusx.MatchInterface(clientInterface),
		dbusx.MatchMembers("LocationUpdated")).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("could not setup D-Bus watch for location updates: %w", err)
	}

	// Create our sensor worker.
	worker.EventType = &locationWorker{
		triggerCh: triggerCh,
		getLocationProperty: func(path, prop string) (float64, error) {
			value, err := dbusx.NewProperty[float64](bus, path, geoclueInterface, locationInterface+"."+prop).Get()
			if err != nil {
				return 0, fmt.Errorf("could not fetch location property %s: %w", prop, err)
			}

			return value, nil
		},
		startMethod: dbusx.NewMethod(bus, geoclueInterface, clientPath, startCall),
		stopMethod:  dbusx.NewMethod(bus, geoclueInterface, clientPath, stopCall),
	}

	return worker, nil
}

func createClient(bus *dbusx.Bus) (string, error) {
	// Check if we can create a client, bail if we can't.
	clientPath, err := dbusx.GetData[string](bus, managerPath, geoclueInterface, getClientCall)
	if clientPath == "" || err != nil {
		return "", fmt.Errorf("could not set up a geoclue client: %w", err)
	}

	// Set an ID for our client.
	if err = dbusx.NewProperty[string](bus, clientPath, geoclueInterface, desktopIDProp).Set(preferences.AppID); err != nil {
		return "", fmt.Errorf("could not set geoclue client id: %w", err)
	}

	return clientPath, nil
}

func setThresholds(bus *dbusx.Bus, clientPath string) {
	var err error

	logger := slog.With(slog.String("worker", workerID))

	// Set a distance threshold.
	if err = dbusx.NewProperty[uint32](bus, clientPath, geoclueInterface, distanceThresholdProp).Set(0); err != nil {
		logger.Debug("Could not set distance threshold for geoclue requests.", slog.Any("error", err))
	}

	// Set a time threshold.
	if err = dbusx.NewProperty[uint32](bus, clientPath, geoclueInterface, timeThresholdProp).Set(0); err != nil {
		logger.Debug("Could not set time threshold for geoclue requests.", slog.Any("error", err))
	}
}
