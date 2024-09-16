// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package apps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"

	workerID = "app_sensors"
)

var ErrNoApps = errors.New("no running apps")

type sensorWorker struct {
	getAppStates func() (map[string]dbus.Variant, error)
	activeApp    *activeAppSensor
	runningApps  *runningAppsSensor
	triggerCh    chan dbusx.Trigger
}

func (w *sensorWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)
	logger := slog.Default().With(slog.String("worker", workerID))

	sendSensors := func(ctx context.Context, sensorCh chan sensor.Details) {
		appSensors, err := w.Sensors(ctx)
		if err != nil {
			logger.Debug("Failed to update app sensors.", slog.Any("error", err))

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

		for range w.triggerCh {
			sendSensors(ctx, sensorCh)
		}
	}()

	return sensorCh, nil
}

func (w *sensorWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	var (
		sensors     []sensor.Details
		runningApps []string
	)

	appStates, err := w.getAppStates()
	if err != nil {
		return nil, err
	}

	for name, variant := range appStates {
		// Convert the state to something we understand.
		state, err := dbusx.VariantToValue[uint32](variant)
		if err != nil {
			continue
		}
		// If the state is greater than 0, this app is running.
		if state > 0 {
			runningApps = append(runningApps, name)
		}
		// If the state is 2 this app is running and the currently active app.
		if state == 2 && w.activeApp.State() != name {
			w.activeApp.Value = name
			sensors = append(sensors, w.activeApp)
		}
	}

	// Update the running apps sensor.
	if w.runningApps.State() != len(runningApps) {
		w.runningApps.Value = len(runningApps)
		w.runningApps.apps = runningApps
		sensors = append(sensors, w.runningApps)
	}

	return sensors, nil
}

func NewAppWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(workerID)

	// If we cannot find a portal interface, we cannot monitor the active app.
	portalDest, ok := linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return worker, linux.ErrNoDesktopPortal
	}

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, linux.ErrNoSessionBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(appStateDBusPath),
		dbusx.MatchInterface(appStateDBusInterface),
		dbusx.MatchMembers("RunningApplicationsChanged"),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("could not watch D-Bus for app state events: %w", err)
	}

	worker.EventType = &sensorWorker{
		triggerCh: triggerCh,
		getAppStates: func() (map[string]dbus.Variant, error) {
			apps, err := dbusx.GetData[map[string]dbus.Variant](bus, appStateDBusPath, portalDest, appStateDBusMethod)
			if err != nil {
				return nil, fmt.Errorf("could not retrieve app list from D-Bus: %w", err)
			}

			if apps == nil {
				return nil, ErrNoApps
			}

			return apps, nil
		},
		activeApp:   newActiveAppSensor(),
		runningApps: newRunningAppsSensor(),
	}

	return worker, nil
}
