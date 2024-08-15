// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package apps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"

	workerID = "app_sensors"
)

type worker struct {
	activeApp   *activeAppSensor
	runningApps *runningAppsSensor
	logger      *slog.Logger
	bus         *dbusx.Bus
	portalDest  string
}

func (w *worker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	triggerCh, err := w.bus.WatchBus(ctx, &dbusx.Watch{
		Path:      appStateDBusPath,
		Interface: appStateDBusInterface,
		Names:     []string{"RunningApplicationsChanged"},
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for app state events: %w", err)
	}

	sendSensors := func(ctx context.Context, sensorCh chan sensor.Details) {
		appSensors, err := w.Sensors(ctx)
		if err != nil {
			w.logger.Warn("Failed to update app sensors.", "error", err.Error())

			return
		}

		for _, s := range appSensors {
			sensorCh <- s
		}
	}

	// Send an initial update.
	go func() {
		sendSensors(ctx, sensorCh)
	}()

	// Listen for and process updates from D-Bus.
	go func() {
		defer close(sensorCh)

		for range triggerCh {
			sendSensors(ctx, sensorCh)
		}
	}()

	return sensorCh, nil
}

func (w *worker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details

	appList, err := dbusx.GetData[map[string]dbus.Variant](ctx, w.bus, appStateDBusPath, w.portalDest, appStateDBusMethod)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve app list from D-Bus: %w", err)
	}

	if appList != nil {
		if s := w.activeApp.update(appList); s != nil {
			sensors = append(sensors, s)
		}

		if s := w.runningApps.update(appList); s != nil {
			sensors = append(sensors, s)
		}
	}

	return sensors, nil
}

func NewAppWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	// If we cannot find a portal interface, we cannot monitor the active app.
	portalDest, err := linux.FindPortal()
	if err != nil {
		return nil, fmt.Errorf("unable to monitor for active applications: %w", err)
	}

	bus, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor for active applications: %w", err)
	}

	return &linux.SensorWorker{
			Value: &worker{
				portalDest:  portalDest,
				activeApp:   newActiveAppSensor(),
				runningApps: newRunningAppsSensor(),
				logger:      logging.FromContext(ctx).With(slog.String("worker", workerID)),
				bus:         bus,
			},
			WorkerID: workerID,
		},
		nil
}
