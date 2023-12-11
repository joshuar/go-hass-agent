// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

type screenlockSensor struct {
	linuxSensor
}

func (s *screenlockSensor) Icon() string {
	state, ok := s.value.(bool)
	if !ok {
		return "mdi:lock-alert"
	}
	if state {
		return "mdi:eye-lock"
	} else {
		return "mdi:eye-lock-open"
	}
}

func (s *screenlockSensor) SensorType() sensor.SensorType {
	return sensor.TypeBinary
}

func newScreenlockEvent(v bool) *screenlockSensor {
	return &screenlockSensor{
		linuxSensor: linuxSensor{
			sensorType: screenLock,
			source:     srcDbus,
			value:      v,
		},
	}
}

func ScreenLockUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	path := dbushelpers.GetSessionPath(ctx)
	if path == "" {
		log.Warn().Msg("Could not ascertain user session from D-Bus. Cannot monitor screen lock state.")
		close(sensorCh)
		return sensorCh
	}
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(path),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Name != dbushelpers.PropChangedSignal || s.Path != path {
				return
			}
			if len(s.Body) <= 1 {
				log.Debug().Caller().Interface("body", s.Body).Msg("Unexpected body length.")
				return
			}
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if !ok {
				log.Debug().Caller().
					Str("signal", s.Name).Interface("body", s.Body).
					Msg("Unexpected signal body")
				return
			}
			if v, ok := props["LockedHint"]; ok {
				sensorCh <- newScreenlockEvent(dbushelpers.VariantToValue[bool](v))
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not poll D-Bus for screen lock. Screen lock sensor will not run.")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped screen lock sensor.")
	}()
	return sensorCh
}
