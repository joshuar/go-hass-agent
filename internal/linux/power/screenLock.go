// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

func newScreenlockEvent(v bool) *screenlockSensor {
	return &screenlockSensor{
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorScreenLock,
			IsBinary:        true,
			SensorSrc:       linux.DataSrcDbus,
			Value:           v,
		},
	}
}

func ScreenLockUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	sessionPath := dbusx.GetSessionPath(ctx)

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{sessionLockSignal, sessionUnlockSignal, sessionLockedProp},
		Interface: sessionInterface,
		Path:      string(sessionPath),
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create screen lock state D-Bus watch.")
		close(sensorCh)
		return sensorCh
	}

	log.Trace().Msg("Started screen lock sensor.")
	go func() {
		defer close(sensorCh)
		for {
			select {
			case <-ctx.Done():
				log.Trace().Msg("Stopped screen lock sensor.")
				return
			case event := <-events:
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
