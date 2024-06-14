// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package apps

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"

	"github.com/rs/zerolog/log"
)

const (
	appStateDBusMethod    = "org.freedesktop.impl.portal.Background.GetAppState"
	appStateDBusPath      = "/org/freedesktop/portal/desktop"
	appStateDBusInterface = "org.freedesktop.impl.portal.Background"
	appStateDBusEvent     = "org.freedesktop.impl.portal.Background.RunningApplicationsChanged"
)

type worker struct {
	activeApp   *activeAppSensor
	runningApps *runningAppsSensor
	portalDest  string
}

//nolint:exhaustruct
func (w *worker) Setup(_ context.Context) (*dbusx.Watch, error) {
	return &dbusx.Watch{
			Bus:       dbusx.SessionBus,
			Path:      appStateDBusPath,
			Interface: appStateDBusInterface,
			Names:     []string{"RunningApplicationsChanged"},
		},
		nil
}

func (w *worker) Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	sendSensors := func(ctx context.Context, sensorCh chan sensor.Details) {
		appSensors, err := w.Sensors(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to update app sensors.")

			return
		}

		for _, s := range appSensors {
			sensorCh <- s
		}
	}

	// Watch for active app changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped app sensor.")

				return
			case <-triggerCh:
				sendSensors(ctx, sensorCh)
			}
		}
	}()

	// Send an initial update.
	go func() {
		sendSensors(ctx, sensorCh)
	}()

	return sensorCh
}

func (w *worker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details

	appList, err := dbusx.GetData[map[string]dbus.Variant](ctx, dbusx.SessionBus, appStateDBusPath, w.portalDest, appStateDBusMethod)
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

func NewAppWorker() (*linux.SensorWorker, error) {
	// If we cannot find a portal interface, we cannot monitor the active app.
	portalDest, err := linux.FindPortal()
	if err != nil {
		return nil, fmt.Errorf("unable to monitor for active applications: %w", err)
	}

	return &linux.SensorWorker{
			WorkerName: "App Sensors",
			WorkerDesc: "Sensors to track the active app and total number of running apps.",
			Value: &worker{
				portalDest:  portalDest,
				activeApp:   newActiveAppSensor(),
				runningApps: newRunningAppsSensor(),
			},
		},
		nil
}
