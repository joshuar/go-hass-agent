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
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	screenLockWorkerID = "screen_lock_sensor"

	screenLockedIcon      = "mdi:eye-lock"
	screenUnlockedIcon    = "mdi:eye-lock-open"
	screenLockUnknownIcon = "mdi:lock-alert"
)

type screenlockSensor struct {
	linux.Sensor
}

func (s *screenlockSensor) Icon() string {
	isLocked, ok := s.Value.(bool)

	switch {
	case !ok:
		return screenLockUnknownIcon
	case isLocked:
		return screenLockedIcon
	default:
		return screenUnlockedIcon
	}
}

func newScreenlockEvent(value bool) *screenlockSensor {
	return &screenlockSensor{
		Sensor: linux.Sensor{
			DisplayName: "Screen Lock",
			IsBinary:    true,
			DataSource:  linux.DataSrcDbus,
			Value:       value,
		},
	}
}

type screenLockWorker struct {
	triggerCh chan dbusx.Trigger
}

func (w *screenLockWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

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
							sensorCh <- newScreenlockEvent(lockState)
						}
					}
				case sessionLockSignal:
					sensorCh <- newScreenlockEvent(true)
				case sessionUnlockSignal:
					sensorCh <- newScreenlockEvent(false)
				}
			}
		}
	}()

	return sensorCh, nil
}

// ?: retrieve the current screen lock state when called.
func (w *screenLockWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

func NewScreenLockWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve D-Bus connection: %w", err)
	}

	sessionPath, err := bus.GetSessionPath()
	if err != nil {
		return nil, fmt.Errorf("unable to determine user session path from D-Bus: %w", err)
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(sessionPath),
		dbusx.MatchMembers(sessionLockSignal, sessionUnlockSignal, sessionLockedProp, "PropertiesChanged"),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to create D-Bus watch for screen lock state: %w", err)
	}

	return &linux.SensorWorker{
			Value: &screenLockWorker{
				triggerCh: triggerCh,
			},
			WorkerID: screenLockWorkerID,
		},
		nil
}
