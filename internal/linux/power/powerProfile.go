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

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	powerProfilesPath      = "/org/freedesktop/UPower/PowerProfiles"
	powerProfilesInterface = "org.freedesktop.UPower.PowerProfiles"
	activeProfileProp      = "ActiveProfile"

	powerProfileWorkerID = "power_profile"
)

func newPowerSensor(profile string) sensor.Entity {
	return sensor.Entity{
		Name:     "Power Profile",
		Category: types.CategoryDiagnostic,
		EntityState: &sensor.EntityState{
			ID:    "power_profile",
			State: profile,
			Icon:  "mdi:flash",
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
			},
		},
	}
}

type profileWorker struct {
	activeProfile *dbusx.Property[string]
	triggerCh     chan dbusx.Trigger
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

func NewProfileWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(powerProfileWorkerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(powerProfilesPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("could not watch D-Bus for power profile updates: %w", err)
	}

	worker.EventType = &profileWorker{
		triggerCh: triggerCh,
		activeProfile: dbusx.NewProperty[string](bus,
			powerProfilesPath, powerProfilesInterface,
			powerProfilesInterface+"."+activeProfileProp),
	}

	return worker, nil
}
