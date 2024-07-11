// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package location

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/device"
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

type locationSensor struct {
	linux.Sensor
}

func (s *locationSensor) Name() string { return "Location" }

func (s *locationSensor) ID() string { return "location" }

type worker struct {
	logger     *slog.Logger
	bus        *dbusx.Bus
	clientPath dbus.ObjectPath
}

//nolint:exhaustruct
func (w *worker) setup(ctx context.Context) (*dbusx.Watch, error) {
	var err error

	// Check if we can create a client, bail if we can't.
	clientPath, err := dbusx.GetData[dbus.ObjectPath](ctx, w.bus, managerPath, geoclueInterface, getClientCall)
	if !clientPath.IsValid() || err != nil {
		return nil, fmt.Errorf("could not set up a geoclue client: %w", err)
	}

	w.clientPath = clientPath

	if err = dbusx.SetProp(ctx, w.bus, string(w.clientPath), geoclueInterface, desktopIDProp, preferences.AppID); err != nil {
		return nil, fmt.Errorf("could not set geoclue client id: %w", err)
	}

	// Set a distance threshold.
	if err = dbusx.SetProp(ctx, w.bus, string(w.clientPath), geoclueInterface, distanceThresholdProp, uint32(0)); err != nil {
		w.logger.Warn("Could not set distance threshold for geoclue requests.", "error", err.Error())
	}

	// Set a time threshold.
	if err = dbusx.SetProp(ctx, w.bus, string(w.clientPath), geoclueInterface, timeThresholdProp, uint32(0)); err != nil {
		w.logger.Warn("Could not set time threshold for geoclue requests.", "error", err.Error())
	}

	// Request to start tracking location updates.
	if err = w.bus.Call(ctx, string(w.clientPath), geoclueInterface, startCall); err != nil {
		return nil, fmt.Errorf("could not start geoclue client: %w", err)
	}

	w.logger.Debug("GeoClue client created.")

	return &dbusx.Watch{
			Path:      string(w.clientPath),
			Interface: clientInterface,
			Names:     []string{"LocationUpdated"},
		},
		nil
}

func (w *worker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	watch, err := w.setup(ctx)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not setup D-Bus watch for location updates: %w", err)
	}

	triggerCh, err := w.bus.WatchBus(ctx, watch)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for location updates: %w", err)
	}

	go func() {
		w.logger.Debug("Monitoring for location updates.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				err := w.bus.Call(ctx, string(w.clientPath), geoclueInterface, stopCall)
				if err != nil {
					w.logger.Debug("Failed to stop location updater.", "error", err.Error())

					return
				}

				return
			case event := <-triggerCh:
				if locationPath, ok := event.Content[1].(dbus.ObjectPath); ok {
					go func() {
						sensorCh <- w.newLocation(ctx, w.logger, locationPath)
					}()
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *worker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

//nolint:exhaustruct
func (w *worker) newLocation(ctx context.Context, logger *slog.Logger, locationPath dbus.ObjectPath) *locationSensor {
	getProp := func(prop string) float64 {
		value, err := dbusx.GetProp[float64](ctx, w.bus, string(locationPath), geoclueInterface, locationInterface+"."+prop)
		if err != nil {
			logger.Debug("Could not retrieve location property.", "property", prop, "error", err.Error())

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

//nolint:exhaustruct
func NewLocationWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	// Don't run this worker if we are not running on a laptop.
	chassis, _ := device.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	if chassis != "laptop" {
		return nil, fmt.Errorf("unable to monitor location updates: %w", device.ErrUnsupportedHardware)
	}

	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor location updates: %w", err)
	}

	return &linux.SensorWorker{
			Value: &worker{
				logger: logging.FromContext(ctx).With(slog.String("worker", workerID)),
				bus:    bus,
			},
			WorkerID: workerID,
		},
		nil
}
