// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
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

func ScreenLockUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace("/org/freedesktop/login1/session"),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), "/org/freedesktop/login1/session") || len(s.Body) <= 1 {
				log.Trace().Caller().Msg("Not my signal or empty signal body.")
				return
			}
			switch s.Name {
			case dbusx.PropChangedSignal:
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					log.Trace().Caller().
						Str("signal", s.Name).Interface("body", s.Body).
						Msg("Unexpected signal body")
					return
				}
				if v, ok := props["LockedHint"]; ok {
					sensorCh <- newScreenlockEvent(dbusx.VariantToValue[bool](v))
				}
				if v, ok := props["IdleHint"]; ok {
					sensorCh <- newScreenlockEvent(dbusx.VariantToValue[bool](v))
				}
			case "org.freedesktop.login1.Session.Lock":
				sensorCh <- newScreenlockEvent(true)
			case "org.freedesktop.login1.Session.Unlock":
				sensorCh <- newScreenlockEvent(false)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not poll D-Bus for screen lock. Screen lock sensor will not run.")
		close(sensorCh)
		return sensorCh
	}
	log.Trace().Msg("Started screen lock sensor.")
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Trace().Msg("Stopped screen lock sensor.")
	}()
	return sensorCh
}
