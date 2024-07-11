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
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	screenLockWorkerID = "screen_lock_sensor"
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

type screenLockWorker struct {
	logger *slog.Logger
	bus    *dbusx.Bus
}

//nolint:cyclop,exhaustruct
func (w *screenLockWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	sessionPath, err := w.bus.GetSessionPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create screen lock worker: %w", err)
	}

	triggerCh, err := w.bus.WatchBus(ctx, &dbusx.Watch{
		Names:     []string{sessionLockSignal, sessionUnlockSignal, sessionLockedProp},
		Interface: sessionInterface,
		Path:      sessionPath,
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for screen lock updates: %w", err)
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				switch event.Signal {
				case dbusx.PropChangedSignal:
					props, err := dbusx.ParsePropertiesChanged(event.Content)
					if err != nil {
						w.logger.Warn("Received unknown event from D-Bus.", "error", err.Error())

						continue
					}

					if lock, lockChanged := props.Changed[sessionLockedProp]; lockChanged {
						if state, err := dbusx.VariantToValue[bool](lock); err != nil {
							w.logger.Warn("Could not screen lock state.", "error", err.Error())
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

func NewScreenLockWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor power state: %w", err)
	}

	return &linux.SensorWorker{
			Value: &screenLockWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", screenLockWorkerID)),
				bus:    bus,
			},
			WorkerID: screenLockWorkerID,
		},
		nil
}
