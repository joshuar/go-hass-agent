// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package desktop

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

var _ workers.EntityWorker = (*sensorWorker)(nil)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"

	appWorkerID   = "app_sensors"
	appWorkerDesc = "App sensors"

	activeAppsIcon = "mdi:application"
	activeAppsName = "Active App"
	activeAppsID   = "active_app"

	runningAppsIcon  = "mdi:apps"
	runningAppsUnits = "apps"
	runningAppsName  = "Running Apps"
	runningAppsID    = "running_apps"
)

func (w *sensorWorker) PreferencesID() string {
	return prefPrefix + "app"
}

func (w *sensorWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{}
}

func (w *sensorWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

type sensorWorker struct {
	bus              *dbusx.Bus
	portalDest       string
	runningApp       string
	totalRunningApps int
	prefs            *WorkerPrefs
	*models.WorkerMetadata
}

func (w *sensorWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(appStateDBusPath),
		dbusx.MatchInterface(appStateDBusInterface),
		dbusx.MatchMembers("RunningApplicationsChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("could not start worker: %w", err)
	}
	sensorCh := make(chan models.Entity)
	logger := slog.Default().With(slog.String("worker", desktopWorkerID))

	sendSensors := func(ctx context.Context, sensorCh chan models.Entity) {
		appSensors, err := w.generateSensors(ctx)
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

		for range triggerCh {
			sendSensors(ctx, sensorCh)
		}
	}()

	return sensorCh, nil
}

func (w *sensorWorker) generateSensors(ctx context.Context) ([]models.Entity, error) {
	var (
		sensors     []models.Entity
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
		if state == 2 && w.runningApp != name {
			w.runningApp = name
			sensors = append(sensors, sensor.NewSensor(ctx,
				sensor.WithName(activeAppsName),
				sensor.WithID(activeAppsID),
				sensor.AsTypeSensor(),
				sensor.WithIcon(activeAppsIcon),
				sensor.WithState(name),
				sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			))
		}
	}

	// Update the running apps sensor.
	if w.totalRunningApps != len(runningApps) {
		w.totalRunningApps = len(runningApps)
		sensors = append(sensors, sensor.NewSensor(ctx,
			sensor.WithName(runningAppsName),
			sensor.WithID(runningAppsID),
			sensor.WithUnits(runningAppsUnits),
			sensor.WithStateClass(class.StateMeasurement),
			sensor.WithIcon(runningAppsIcon),
			sensor.WithState(w.totalRunningApps),
			sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			sensor.WithAttribute("apps", runningApps),
		))
	}

	return sensors, nil
}

func (w *sensorWorker) getAppStates() (map[string]dbus.Variant, error) {
	var apps map[string]dbus.Variant
	apps, err := dbusx.GetData[map[string]dbus.Variant](w.bus, appStateDBusPath, w.portalDest, appStateDBusMethod)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve app states from D-Bus: %w", err)
	}
	return apps, nil
}

// NewAppStateWorker creates a worker for tracking application states (i.e., number of running and current active apps).
func NewAppStateWorker(ctx context.Context) (workers.EntityWorker, error) {
	// If we cannot find a portal interface, we cannot monitor the active app.
	portalDest, ok := linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return nil, fmt.Errorf("could not start apps worker: %w", linux.ErrNoDesktopPortal)
	}

	// Connect to the D-Bus session bus. Bail if we can't.
	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, linux.ErrNoSessionBus
	}

	worker := &sensorWorker{
		WorkerMetadata: models.SetWorkerMetadata(appWorkerID, appWorkerDesc),
		bus:            bus,
		portalDest:     portalDest,
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return worker, fmt.Errorf("could not start apps worker: %w", err)
	}
	worker.prefs = prefs

	return worker, nil
}
