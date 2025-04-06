// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
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

var (
	ErrNewPowerProfileSensor  = errors.New("could not create power profile sensor")
	ErrInitPowerProfileWorker = errors.New("could not init power profile worker")
)

func newPowerSensor(ctx context.Context, profile string) (*models.Entity, error) {
	profileSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Power Profile"),
		sensor.WithID("power_profile"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:flash"),
		sensor.WithState(profile),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
	)
	if err != nil {
		return nil, errors.Join(ErrNewPowerProfileSensor, err)
	}

	return &profileSensor, nil
}

type profileWorker struct {
	bus           *dbusx.Bus
	activeProfile *dbusx.Property[string]
	logger        *slog.Logger
	prefs         *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

//nolint:gocognit
func (w *profileWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(powerProfilesPath),
		dbusx.MatchPropChanged(),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, errors.Join(ErrInitPowerProfileWorker,
			fmt.Errorf("could not watch D-Bus for power profile updates: %w", err))
	}
	sensorCh := make(chan models.Entity)

	// Get the current power profile and send it as an initial sensor value.
	sensors, err := w.Sensors(ctx)
	if err != nil {
		w.logger.Debug("Could not retrieve power profile.", slog.Any("error", err))
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
			case event := <-triggerCh:
				changed, profile, err := dbusx.HasPropertyChanged[string](event.Content, activeProfileProp)
				if err != nil {
					w.logger.Debug("Could not parse received D-Bus signal.", slog.Any("error", err))
				} else {
					if changed {
						entity, err := newPowerSensor(ctx, profile)
						if err != nil {
							w.logger.Warn("Could not generate power profile sensor.", slog.Any("error", err))
						} else {
							sensorCh <- *entity
						}
					}
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *profileWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	profile, err := w.activeProfile.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve active power profile from D-Bus: %w", err)
	}

	entity, err := newPowerSensor(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("unable to generate active power profile sensor: %w", err)
	}

	return []models.Entity{*entity}, nil
}

func (w *profileWorker) PreferencesID() string {
	return powerProfilePreferencesID
}

func (w *profileWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *profileWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func NewProfileWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitPowerProfileWorker, linux.ErrNoSystemBus)
	}

	worker := &profileWorker{
		WorkerMetadata: models.SetWorkerMetadata(powerProfileWorkerID, powerProfileWorkerDesc),
		bus:            bus,
		logger:         slog.With(slog.String("worker", powerProfileWorkerID)),

		activeProfile: dbusx.NewProperty[string](bus,
			powerProfilesPath, powerProfilesInterface,
			powerProfilesInterface+"."+activeProfileProp),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitPowerProfileWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
