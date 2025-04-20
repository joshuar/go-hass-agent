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

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	screenLockWorkerID      = "screen_lock_sensor"
	screenLockWorkerDesc    = "Screen lock detection"
	screenLockPreferencesID = sensorsPrefPrefix + "screen_lock"

	screenLockedIcon      = "mdi:eye-lock"
	screenUnlockedIcon    = "mdi:eye-lock-open"
	screenLockUnknownIcon = "mdi:lock-alert"
)

var _ workers.EntityWorker = (*screenLockWorker)(nil)

var (
	ErrNewScreenLockSensor  = errors.New("could not create screen lock sensor")
	ErrInitScreenLockWorker = errors.New("could not init screen lock worker")
)

func newScreenlockSensor(ctx context.Context, value bool) (models.Entity, error) {
	lockSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Screen Lock"),
		sensor.WithID("screen_lock"),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(class.BinaryClassLock),
		sensor.WithIcon(screenLockIcon(value)),
		sensor.WithState(!value), // For device class BinarySensorDeviceClassLock: On means open (unlocked), Off means closed (locked).
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
		sensor.AsRetryableRequest(true),
	)
	if err != nil {
		return models.Entity{}, errors.Join(ErrNewScreenLockSensor, err)
	}

	return lockSensor, nil
}

func screenLockIcon(value bool) string {
	switch value {
	case true:
		return screenLockedIcon
	default:
		return screenUnlockedIcon
	}
}

type screenLockWorker struct {
	bus            *dbusx.Bus
	sessionPath    string
	screenLockProp *dbusx.Property[bool]
	prefs          *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

//nolint:gocognit
func (w *screenLockWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(w.sessionPath),
		dbusx.MatchMembers(sessionLockSignal, sessionUnlockSignal, sessionLockedProp, "PropertiesChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, errors.Join(ErrInitScreenLockWorker,
			fmt.Errorf("unable to create D-Bus watch for screen lock state: %w", err))
	}
	sensorCh := make(chan models.Entity)

	currentState, err := w.getCurrentState(ctx)
	if err != nil {
		close(sensorCh)
		return sensorCh, fmt.Errorf("cannot process screen lock events: %w", err)
	}

	go func() {
		sensorCh <- currentState
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(sensorCh)
				return
			case event := <-triggerCh:
				var (
					entity    models.Entity
					lockState bool
					changed   bool
					err       error
				)

				switch event.Signal {
				case dbusx.PropChangedSignal:
					changed, lockState, err = dbusx.HasPropertyChanged[bool](event.Content, sessionLockedProp)
					if err != nil {
						slog.With(slog.String("worker", screenLockWorkerID)).Debug("Could not parse received D-Bus signal.", slog.Any("error", err))
					} else {
						if changed {
							entity, err = newScreenlockSensor(ctx, lockState)
						}
					}
				case sessionLockSignal:
					entity, err = newScreenlockSensor(ctx, true)
				case sessionUnlockSignal:
					entity, err = newScreenlockSensor(ctx, false)
				}

				if err != nil {
					slogctx.FromCtx(ctx).Warn("Could not generate screen lock sensor.",
						slog.Any("error", err))
					continue
				}

				sensorCh <- entity
			}
		}
	}()

	return sensorCh, nil
}

func (w *screenLockWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	currentState, err := w.getCurrentState(ctx)
	if err != nil {
		return nil, errors.Join(ErrNewScreenLockSensor, err)
	}

	return []models.Entity{currentState}, nil
}

func (w *screenLockWorker) PreferencesID() string {
	return screenLockPreferencesID
}

func (w *screenLockWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *screenLockWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *screenLockWorker) getCurrentState(ctx context.Context) (models.Entity, error) {
	screenLockState, err := w.screenLockProp.Get()
	if err != nil {
		return models.Entity{}, errors.Join(ErrNewScreenLockSensor, err)
	}

	return newScreenlockSensor(ctx, screenLockState)
}

func NewScreenLockWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitScreenLockWorker, linux.ErrNoSystemBus)
	}

	sessionPath, ok := linux.CtxGetSessionPath(ctx)
	if !ok {
		return nil, errors.Join(ErrInitScreenLockWorker, linux.ErrNoSessionPath)
	}

	worker := &screenLockWorker{
		WorkerMetadata: models.SetWorkerMetadata(screenLockWorkerID, screenLockWorkerDesc),
		bus:            bus,
		sessionPath:    sessionPath,
		screenLockProp: dbusx.NewProperty[bool](bus, sessionPath, loginBaseInterface, sessionInterface+"."+sessionLockedProp),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitScreenLockWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
