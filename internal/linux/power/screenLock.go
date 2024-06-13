// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

//nolint:exhaustruct
func (w *screenLockWorker) Setup(ctx context.Context) *dbusx.Watch {
	sessionPath := dbusx.GetSessionPath(ctx)

	return &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sessionLockSignal, sessionUnlockSignal, sessionLockedProp},
		Interface: sessionInterface,
		Path:      string(sessionPath),
	}
}

func (w *screenLockWorker) Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Trace().Msg("Stopped screen lock sensor.")

				return
			case event := <-triggerCh:
				switch event.Signal {
				case dbusx.PropChangedSignal:
					props, err := dbusx.ParsePropertiesChanged(event.Content)
					if err != nil {
						log.Warn().Err(err).Msg("Did not understand received trigger.")

						continue
					}

					if state, lockChanged := props.Changed[sessionLockedProp]; lockChanged {
						sensorCh <- newScreenlockEvent(dbusx.VariantToValue[bool](state))
					}
				case sessionLockSignal:
					sensorCh <- newScreenlockEvent(true)
				case sessionUnlockSignal:
					sensorCh <- newScreenlockEvent(false)
				}
			}
		}
	}()

	return sensorCh
}

// ?: retrieve the current screen lock state when called.
func (w *screenLockWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

func NewScreenLockWorker(_ context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Screen Lock Sensor",
			WorkerDesc: "Sensor to track whether the screen is currently locked.",
			Value:      &screenLockWorker{},
		},
		nil
}
