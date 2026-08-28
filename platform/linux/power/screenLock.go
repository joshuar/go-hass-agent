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
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	screenLockedIcon      = "mdi:eye-lock"
	screenUnlockedIcon    = "mdi:eye-lock-open"
	screenLockUnknownIcon = "mdi:lock-alert"
)

var _ workers.EntityWorker = (*screenLockWorker)(nil)

func newScreenlockSensor(ctx context.Context, value bool) models.Entity {
	return models.NewSensor(
		ctx,
		models.WithName("Screen Lock"),
		models.WithID("screen_lock"),
		models.AsTypeBinarySensor(),
		models.WithDeviceClass(models.BinaryClassLock),
		models.WithIcon(screenLockIcon(value)),
		models.WithState(
			!value,
		), // For device class BinarySensorDeviceClassLock: On means open (unlocked), Off means closed (locked).
		models.WithDataSourceAttribute(linux.DataSrcDBus),
		models.AsRetryableRequest(true),
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
	*models.WorkerMetadata

	bus            *dbusx.Bus
	sessionPath    string
	screenLockProp *dbusx.Property[bool]
	prefs          *workers.CommonWorkerPrefs
}

func NewScreenLockWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &screenLockWorker{
		WorkerMetadata: models.SetWorkerMetadata("screen_lock", "Screen lock"),
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}

	// The session path is absent when the user has no graphical session. Rather
	// than not reporting at all, the worker then reports the screen as locked:
	// without a graphical session, the desktop is not accessible. The state is
	// re-evaluated when the agent restarts, which is what happens when a
	// graphical session starts or ends.
	worker.sessionPath, ok = linux.CtxGetSessionPath(ctx)
	if ok {
		worker.screenLockProp =
			dbusx.NewProperty[bool](
				worker.bus,
				worker.sessionPath,
				loginBaseInterface,
				sessionInterface+"."+sessionLockedProp,
			)
	} else {
		slogctx.FromCtx(ctx).Debug("No graphical session, screen lock will be reported as locked.")
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(sensorsPrefPrefix+"screen_lock", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *screenLockWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	// Without a graphical session there is nothing to watch. Report the screen
	// as locked and leave it at that.
	if w.sessionPath == "" {
		sensorCh := make(chan models.Entity)

		go func() {
			defer close(sensorCh)

			select {
			case sensorCh <- newScreenlockSensor(ctx, true):
			case <-ctx.Done():
				return
			}

			<-ctx.Done()
		}()

		return sensorCh, nil
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(w.sessionPath),
		dbusx.MatchMembers(sessionLockSignal, sessionUnlockSignal, sessionLockedProp, "PropertiesChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch screen lock: %w", err)
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
					switch {
					case err != nil:
						slogctx.FromCtx(ctx).Debug("Could not parse received D-Bus signal.", slog.Any("error", err))
					case changed:
						sensorCh <- newScreenlockSensor(ctx, lockState)
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
