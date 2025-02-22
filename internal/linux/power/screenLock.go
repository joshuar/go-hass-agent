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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	screenLockWorkerID      = "screen_lock_sensor"
	screenLockPreferencesID = sensorsPrefPrefix + "screen_lock"

	screenLockedIcon      = "mdi:eye-lock"
	screenUnlockedIcon    = "mdi:eye-lock-open"
	screenLockUnknownIcon = "mdi:lock-alert"
)

var (
	ErrNewScreenLockSensor  = errors.New("could not create screen lock sensor")
	ErrInitScreenLockWorker = errors.New("could not init screen lock worker")
)

func newScreenlockSensor(ctx context.Context, value bool) (*models.Entity, error) {
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
		return nil, errors.Join(ErrNewScreenLockSensor, err)
	}

	return &lockSensor, nil
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
	triggerCh      chan dbusx.Trigger
	screenLockProp *dbusx.Property[bool]
	prefs          *preferences.CommonWorkerPrefs
}

//nolint:gocognit
func (w *screenLockWorker) Events(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

	currentState, err := w.getCurrentState(ctx)
	if err != nil {
		close(sensorCh)
		return sensorCh, fmt.Errorf("cannot process screen lock events: %w", err)
	}

	go func() {
		sensorCh <- *currentState
	}()

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				var (
					entity    *models.Entity
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
					logging.FromContext(ctx).Warn("Could not generate screen lock sensor.",
						slog.Any("error", err))
					continue
				}

				sensorCh <- *entity
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

	return []models.Entity{*currentState}, nil
}

func (w *screenLockWorker) PreferencesID() string {
	return screenLockPreferencesID
}

func (w *screenLockWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *screenLockWorker) getCurrentState(ctx context.Context) (*models.Entity, error) {
	screenLockState, err := w.screenLockProp.Get()
	if err != nil {
		return nil, errors.Join(ErrNewScreenLockSensor, err)
	}

	return newScreenlockSensor(ctx, screenLockState)
}

func NewScreenLockWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(screenLockWorkerID)
	lockWorker := &screenLockWorker{}

	lockWorker.prefs, err = preferences.LoadWorker(lockWorker)
	if err != nil {
		return nil, errors.Join(ErrInitScreenLockWorker, err)
	}

	if lockWorker.prefs.IsDisabled() {
		return worker, nil
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, errors.Join(ErrInitScreenLockWorker, linux.ErrNoSystemBus)
	}

	sessionPath, ok := linux.CtxGetSessionPath(ctx)
	if !ok {
		return worker, errors.Join(ErrInitScreenLockWorker, linux.ErrNoSessionPath)
	}

	lockWorker.triggerCh, err = dbusx.NewWatch(
		dbusx.MatchPath(sessionPath),
		dbusx.MatchMembers(sessionLockSignal, sessionUnlockSignal, sessionLockedProp, "PropertiesChanged"),
	).Start(ctx, bus)
	if err != nil {
		return worker, errors.Join(ErrInitScreenLockWorker,
			fmt.Errorf("unable to create D-Bus watch for screen lock state: %w", err))
	}

	lockWorker.screenLockProp = dbusx.NewProperty[bool](bus, sessionPath, loginBaseInterface, sessionInterface+"."+sessionLockedProp)

	worker.EventSensorType = lockWorker

	return worker, nil
}
