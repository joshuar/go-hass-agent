// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package desktop

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*appsWorker)(nil)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"

	activeAppsIcon = "mdi:application"
	activeAppsName = "Active App"
	activeAppsID   = "active_app"

	runningAppsIcon  = "mdi:apps"
	runningAppsUnits = "apps"
	runningAppsName  = "Running Apps"
	runningAppsID    = "running_apps"
)

type appsWorker struct {
	*models.WorkerMetadata

	bus              *dbusx.Bus
	portalDest       string
	runningApp       string
	totalRunningApps int
	prefs            *WorkerPrefs
}

// NewAppStateWorker creates a worker for tracking application states (i.e., number of running and current active apps).
func NewAppStateWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &appsWorker{
		WorkerMetadata: models.SetWorkerMetadata("running_apps", "Running apps"),
	}

	var ok bool

	// If we cannot find a portal interface, we cannot monitor the active app.
	worker.portalDest, ok = linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return worker, fmt.Errorf("get desktop portal: %w", linux.ErrNoDesktopPortal)
	}

	// Connect to the D-Bus session bus. Bail if we can't.
	worker.bus, ok = linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, linux.ErrNoSessionBus
	}

	defaultPrefs := &WorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(prefPrefix+"app_sensors", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *appsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(appStateDBusPath),
		dbusx.MatchInterface(appStateDBusInterface),
		dbusx.MatchMembers("RunningApplicationsChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch app states: %w", err)
	}
	sensorCh := make(chan models.Entity)

	sendSensors := func(ctx context.Context, sensorCh chan models.Entity) {
		appSensors, err := w.generateSensors(ctx)
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Failed to update app sensors.", slog.Any("error", err))

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

func (w *appsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *appsWorker) generateSensors(ctx context.Context) ([]models.Entity, error) {
	var (
		sensors     []models.Entity
		runningApps []string
	)

	var apps map[string]dbus.Variant
	apps, err := dbusx.GetData[map[string]dbus.Variant](w.bus, appStateDBusPath, w.portalDest, appStateDBusMethod)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve app states from D-Bus: %w", err)
	}

	for name, variant := range apps {
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
				sensor.WithDataSourceAttribute(linux.DataSrcDBus),
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
			sensor.WithDataSourceAttribute(linux.DataSrcDBus),
			sensor.WithAttribute("apps", runningApps),
		))
	}

	return sensors, nil
}
