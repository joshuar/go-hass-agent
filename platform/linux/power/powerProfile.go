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

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	powerProfilesPath      = "/org/freedesktop/UPower/PowerProfiles"
	powerProfilesInterface = "org.freedesktop.UPower.PowerProfiles"
	activeProfileProp      = "ActiveProfile"

	powerProfileWorkerID      = "power_profile_sensor"
	powerProfileWorkerDesc    = "Active power profile"
	powerProfilePreferencesID = sensorsPrefPrefix + "profile"
)

var _ workers.EntityWorker = (*profileWorker)(nil)

func newPowerSensor(ctx context.Context, profile string) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName("Power Profile"),
		sensor.WithID("power_profile"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:flash"),
		sensor.WithState(profile),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
	)
}

type profileWorker struct {
	*models.WorkerMetadata

	bus           *dbusx.Bus
	activeProfile *dbusx.Property[string]
	prefs         *workers.CommonWorkerPrefs
}

func NewProfileWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &profileWorker{
		WorkerMetadata: models.SetWorkerMetadata(powerProfileWorkerID, powerProfileWorkerDesc),
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}
	worker.activeProfile = dbusx.NewProperty[string](worker.bus,
		powerProfilesPath, powerProfilesInterface,
		powerProfilesInterface+"."+activeProfileProp)

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(powerProfilePreferencesID, defaultPrefs)
	if err != nil {
		return nil, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *profileWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(powerProfilesPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch power profile status: %w", err)
	}
	sensorCh := make(chan models.Entity)

	// Get the current power profile and send it as an initial sensor value.
	profile, err := w.activeProfile.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve active power profile from D-Bus: %w", err)
	}
	go func() {
		sensorCh <- newPowerSensor(ctx, profile)
	}()

	// Watch for power profile changes.
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				changed, powerProfile, err := dbusx.HasPropertyChanged[string](event.Content, activeProfileProp)
				switch {
				case err != nil:
					slogctx.FromCtx(ctx).Debug("Could not parse received D-Bus signal.", slog.Any("error", err))
				case changed:
					sensorCh <- newPowerSensor(ctx, powerProfile)
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *profileWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
