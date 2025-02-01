// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	powerProfilesPath      = "/org/freedesktop/UPower/PowerProfiles"
	powerProfilesInterface = "org.freedesktop.UPower.PowerProfiles"
	activeProfileProp      = "ActiveProfile"

	powerProfileWorkerID      = "power_profile_sensor"
	powerProfilePreferencesID = sensorsPrefPrefix + "profile"
)

func newPowerSensor(profile string) sensor.Entity {
	return sensor.NewSensor(
		sensor.WithName("Power Profile"),
		sensor.WithID("power_profile"),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon("mdi:flash"),
			sensor.WithValue(profile),
			sensor.WithDataSourceAttribute(linux.DataSrcDbus),
		),
	)
}

type profileWorker struct {
	activeProfile *dbusx.Property[string]
	triggerCh     chan dbusx.Trigger
	prefs         *preferences.CommonWorkerPrefs
}

func (w *profileWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)
	logger := slog.With(slog.String("worker", powerProfileWorkerID))

	// Get the current power profile and send it as an initial sensor value.
	sensors, err := w.Sensors(ctx)
	if err != nil {
		logger.Debug("Could not retrieve power profile.", slog.Any("error", err))
	} else {
		go func() {
			sensorCh <- sensors[0]
		}()
	}

	// Watch for power profile changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				changed, profile, err := dbusx.HasPropertyChanged[string](event.Content, activeProfileProp)
				if err != nil {
					logger.Debug("Could not parse received D-Bus signal.", slog.Any("error", err))
				} else {
					if changed {
						sensorCh <- newPowerSensor(profile)
					}
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *profileWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	profile, err := w.activeProfile.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve active power profile from D-Bus: %w", err)
	}

	return []sensor.Entity{newPowerSensor(profile)}, nil
}

func (w *profileWorker) PreferencesID() string {
	return powerProfilePreferencesID
}

func (w *profileWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewProfileWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(powerProfileWorkerID)
	powerProfileWorker := &profileWorker{}

	powerProfileWorker.prefs, err = preferences.LoadWorker(powerProfileWorker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	if powerProfileWorker.prefs.IsDisabled() {
		return worker, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	powerProfileWorker.triggerCh, err = dbusx.NewWatch(
		dbusx.MatchPath(powerProfilesPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("could not watch D-Bus for power profile updates: %w", err)
	}

	powerProfileWorker.activeProfile = dbusx.NewProperty[string](bus,
		powerProfilesPath, powerProfilesInterface,
		powerProfilesInterface+"."+activeProfileProp)

	worker.EventSensorType = powerProfileWorker

	return worker, nil
}
