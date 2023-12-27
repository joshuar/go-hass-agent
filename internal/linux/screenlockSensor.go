// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
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
	}
	return "mdi:eye-lock-open"
}

func newScreenlockEvent(v bool) *screenlockSensor {
	return &screenlockSensor{
		linuxSensor: linuxSensor{
			sensorType: screenLock,
			isBinary:   true,
			source:     srcDbus,
			value:      v,
		},
	}
}

func ScreenLockUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace("/org/freedesktop/login1/session"),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), "/org/freedesktop/login1/session") || len(s.Body) <= 1 {
				log.Trace().Caller().Msg("Not my signal or empty signal body.")
				return
			}
			switch s.Name {
			case dbushelpers.PropChangedSignal:
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					log.Trace().Caller().
						Str("signal", s.Name).Interface("body", s.Body).
						Msg("Unexpected signal body")
					return
				}
				if v, ok := props["LockedHint"]; ok {
					sensorCh <- newScreenlockEvent(dbushelpers.VariantToValue[bool](v))
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
