// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type screenlockSensor struct {
	linux.Sensor
}

func (s *screenlockSensor) Icon() string {
	state, ok := s.Value.(bool)
	if !ok {
		return "mdi:lock-alert"
	}

	if state {
		return "mdi:eye-lock"
	}

	return "mdi:eye-lock-open"
}

//nolint:exhaustruct
func newScreenlockEvent(value bool) *screenlockSensor {
	return &screenlockSensor{
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorScreenLock,
			IsBinary:        true,
			SensorSrc:       linux.DataSrcDbus,
			Value:           value,
		},
	}
}

type screenLockWorker struct{}

//nolint:cyclop,exhaustruct
func (w *screenLockWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	sessionPath, err := dbusx.GetSessionPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create screen lock worker: %w", err)
	}

	triggerCh, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sessionLockSignal, sessionUnlockSignal, sessionLockedProp},
		Interface: sessionInterface,
		Path:      sessionPath,
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for screen lock updates: %w", err)
	}

	go func() {
		logging.FromContext(ctx).Debug("Monitoring screen lock.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				logging.FromContext(ctx).Debug("Unmonitoring screen lock.")

				return
			case event := <-triggerCh:
				switch event.Signal {
				case dbusx.PropChangedSignal:
					props, err := dbusx.ParsePropertiesChanged(event.Content)
					if err != nil {
						logging.FromContext(ctx).Warn("Received unknown event from D-Bus.", "error", err.Error())

						continue
					}

					if lock, lockChanged := props.Changed[sessionLockedProp]; lockChanged {
						if state, err := dbusx.VariantToValue[bool](lock); err != nil {
							logging.FromContext(ctx).Warn("Could not screen lock state.", "error", err.Error())
						} else {
							sensorCh <- newScreenlockEvent(state)
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

func NewScreenLockWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Screen Lock Sensor",
			WorkerDesc: "Sensor to track whether the screen is currently locked.",
			Value:      &screenLockWorker{},
		},
		nil
}
