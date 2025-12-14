// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package location

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/location"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*locationWorker)(nil)

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

	workerID      = "location_worker"
	workerDesc    = "Location"
	preferencesID = "sensors.location"
)

type locationWorker struct {
	*models.WorkerMetadata

	bus   *dbusx.Bus
	prefs *workers.CommonWorkerPrefs
}

func NewLocationWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &locationWorker{
		WorkerMetadata: models.SetWorkerMetadata(workerID, workerDesc),
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}

	// Load the worker preferences.
	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(preferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

//nolint:gocognit
func (w *locationWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	// Create a GeoClue client.
	clientPath, err := w.createClient()
	if err != nil {
		return nil, fmt.Errorf("create geoclue client: %w", err)
	}
	// Set start/stop methods.
	startMethod := dbusx.NewMethod(w.bus, geoclueInterface, clientPath, startCall)
	stopMethod := dbusx.NewMethod(w.bus, geoclueInterface, clientPath, stopCall)
	// Set threshold values.
	w.setThresholds(ctx, clientPath)
	// Start GeoClue client.
	if err := startMethod.Call(ctx); err != nil {
		return nil, fmt.Errorf("could not start geoclue client: %w", err)
	}
	// Watch for location changes.
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(clientPath),
		dbusx.MatchInterface(clientInterface),
		dbusx.MatchMembers("LocationUpdated")).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch for location updates: %w", err)
	}

	sensorCh := make(chan models.Entity)

	go func() {
		slogctx.FromCtx(ctx).Debug("Monitoring for location updates.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				if err := stopMethod.Call(ctx); err != nil {
					slogctx.FromCtx(ctx).Debug("Could not stop geoclue client.", slog.Any("error", err))
				}

				return
			case event := <-triggerCh:
				if locationPath, ok := event.Content[1].(dbus.ObjectPath); ok {
					go func() {
						locationSensor, err := w.newLocation(ctx, string(locationPath))
						if err != nil {
							slogctx.FromCtx(ctx).Error("Could not update location.", slog.Any("error", err))
						} else {
							slogctx.FromCtx(ctx).Debug("Location updated.")
							sensorCh <- locationSensor
						}
					}()
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *locationWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *locationWorker) newLocation(ctx context.Context, locationPath string) (models.Entity, error) {
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

	return location.NewLocation(ctx,
		location.WithGPSCoords(float32(latitude), float32(longitude)),
		location.WithGPSAccuracy(int(accuracy)),
		location.WithSpeed(int(speed)),
		location.WithAltitude(int(altitude)),
	), warnings
}

func (w *locationWorker) createClient() (string, error) {
	// Check if we can create a client, bail if we can't.
	clientPath, err := dbusx.GetData[string](w.bus, managerPath, geoclueInterface, getClientCall)
	if clientPath == "" || err != nil {
		return "", fmt.Errorf("could not set up a geoclue client: %w", err)
	}

	// Set an ID for our client.
	if err = dbusx.NewProperty[string](w.bus, clientPath, geoclueInterface, desktopIDProp).Set(config.AppID); err != nil {
		return "", fmt.Errorf("could not set geoclue client id: %w", err)
	}

	return clientPath, nil
}

func (w *locationWorker) setThresholds(ctx context.Context, clientPath string) {
	// Set a distance threshold.
	if err := dbusx.NewProperty[uint32](w.bus, clientPath, geoclueInterface, distanceThresholdProp).Set(0); err != nil {
		slogctx.FromCtx(ctx).Debug("Could not set distance threshold for geoclue requests.", slog.Any("error", err))
	}
	// Set a time threshold.
	if err := dbusx.NewProperty[uint32](w.bus, clientPath, geoclueInterface, timeThresholdProp).Set(0); err != nil {
		slogctx.FromCtx(ctx).Debug("Could not set time threshold for geoclue requests.", slog.Any("error", err))
	}
}

func (w *locationWorker) getLocationProperty(path, prop string) (float64, error) {
	value, err := dbusx.NewProperty[float64](w.bus, path, geoclueInterface, locationInterface+"."+prop).Get()
	if err != nil {
		return 0, fmt.Errorf("get location property: %w", err)
	}

	return value, nil
}
