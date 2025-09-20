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

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
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

func newScreenlockSensor(ctx context.Context, value bool) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName("Screen Lock"),
		sensor.WithID("screen_lock"),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(class.BinaryClassLock),
		sensor.WithIcon(screenLockIcon(value)),
		sensor.WithState(!value), // For device class BinarySensorDeviceClassLock: On means open (unlocked), Off means closed (locked).
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
		sensor.AsRetryableRequest(true),
	)
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
	prefs          *workers.CommonWorkerPrefs
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

	screenLockState, err := w.screenLockProp.Get()
	if err != nil {
		close(sensorCh)
		return sensorCh, fmt.Errorf("cannot process screen lock events: %w", err)
	}

	// Send an initial update.
	go func() {
		sensorCh <- newScreenlockSensor(ctx, screenLockState)
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(sensorCh)
				return
			case event := <-triggerCh:
				var (
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
							sensorCh <- newScreenlockSensor(ctx, lockState)
						}
					}
				case sessionLockSignal:
					sensorCh <- newScreenlockSensor(ctx, true)
				case sessionUnlockSignal:
					sensorCh <- newScreenlockSensor(ctx, false)
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *screenLockWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
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

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(screenLockPreferencesID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitScreenLockWorker, err)
	}

	return worker, nil
}
