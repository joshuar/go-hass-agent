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
	screenLockWorkerID = "screen_lock_sensor"

	screenLockedIcon      = "mdi:eye-lock"
	screenUnlockedIcon    = "mdi:eye-lock-open"
	screenLockUnknownIcon = "mdi:lock-alert"
)

func newScreenlockSensor(value bool) sensor.Entity {
	return sensor.Entity{
		Name: "Screen Lock",
		EntityState: &sensor.EntityState{
			ID:         "screen_lock",
			Icon:       screenLockIcon(value),
			EntityType: types.BinarySensor,
			State:      value,
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
			},
		},
	}
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
	triggerCh chan dbusx.Trigger
}

func (w *screenLockWorker) Events(ctx context.Context) (chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				switch event.Signal {
				case dbusx.PropChangedSignal:
					changed, lockState, err := dbusx.HasPropertyChanged[bool](event.Content, sessionLockedProp)
					if err != nil {
						slog.With(slog.String("worker", screenLockWorkerID)).Debug("Could not parse received D-Bus signal.", slog.Any("error", err))
					} else {
						if changed {
							sensorCh <- newScreenlockSensor(lockState)
						}
					}
				case sessionLockSignal:
					sensorCh <- newScreenlockSensor(true)
				case sessionUnlockSignal:
					sensorCh <- newScreenlockSensor(false)
				}
			}
		}
	}()

	return sensorCh, nil
}

// ?: retrieve the current screen lock state when called.
func (w *screenLockWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return nil, linux.ErrUnimplemented
}

func NewScreenLockWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(screenLockWorkerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	sessionPath, ok := linux.CtxGetSessionPath(ctx)
	if !ok {
		return worker, linux.ErrNoSessionPath
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(sessionPath),
		dbusx.MatchMembers(sessionLockSignal, sessionUnlockSignal, sessionLockedProp, "PropertiesChanged"),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("unable to create D-Bus watch for screen lock state: %w", err)
	}

	worker.EventType = &screenLockWorker{
		triggerCh: triggerCh,
	}

	return worker, nil
}
