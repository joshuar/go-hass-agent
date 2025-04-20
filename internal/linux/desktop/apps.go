// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package desktop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

var _ workers.EntityWorker = (*sensorWorker)(nil)

var (
	ErrInitAppsWorker = errors.New("could not init apps worker")
	ErrNoApps         = errors.New("no running apps")
)

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
	triggerCh        <-chan dbusx.Trigger
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
		return nil, errors.Join(ErrInitAppsWorker, fmt.Errorf("could not watch D-Bus for app state events: %w", err))
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

			entity, err := sensor.NewSensor(ctx,
				sensor.WithName(activeAppsName),
				sensor.WithID(activeAppsID),
				sensor.AsTypeSensor(),
				sensor.WithIcon(activeAppsIcon),
				sensor.WithState(name),
				sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			)
			if err != nil {
				slogctx.FromCtx(ctx).Warn("Could not generate active app sensor.", slog.Any("error", err))
			} else {
				sensors = append(sensors, entity)
			}
		}
	}

	// Update the running apps sensor.
	if w.totalRunningApps != len(runningApps) {
		entity, err := sensor.NewSensor(ctx,
			sensor.WithName(runningAppsName),
			sensor.WithID(runningAppsID),
			sensor.WithUnits(runningAppsUnits),
			sensor.WithStateClass(class.StateMeasurement),
			sensor.WithIcon(runningAppsIcon),
			sensor.WithState(len(runningApps)),
			sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			sensor.WithAttribute("apps", runningApps),
		)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not generate active app sensor.", slog.Any("error", err))
		} else {
			sensors = append(sensors, entity)
		}

		w.totalRunningApps = len(runningApps)
	}

	return sensors, nil
}

func (w *sensorWorker) getAppStates() (map[string]dbus.Variant, error) {
	var apps map[string]dbus.Variant
	apps, err := dbusx.GetData[map[string]dbus.Variant](w.bus, appStateDBusPath, w.portalDest, appStateDBusMethod)
	if err != nil {
		return nil, errors.Join(ErrInitAppsWorker, fmt.Errorf("could not retrieve app list from D-Bus: %w", err))
	}

	if apps == nil {
		return nil, errors.Join(ErrInitAppsWorker, ErrNoApps)
	}

	return apps, nil
}

func NewAppWorker(ctx context.Context) (workers.EntityWorker, error) {
	// If we cannot find a portal interface, we cannot monitor the active app.
	portalDest, ok := linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return nil, errors.Join(ErrInitAppsWorker, linux.ErrNoDesktopPortal)
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
		return worker, errors.Join(ErrInitAppsWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
