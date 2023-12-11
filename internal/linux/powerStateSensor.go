// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

type powerStateSensor struct {
	linuxSensor
}

func (s *powerStateSensor) Icon() string {
	state, ok := s.value.(string)
	if !ok {
		return "mdi:power-on"
	} else {
		switch state {
		case "Suspended":
			return "mdi:power-sleep"
		case "Powered Off":
			return "mdi:power-off"
		default:
			return "mdi:power-on"
		}
	}
}

func PowerStateUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)

	sensorCh <- newPowerState("Powered On")

	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath("/org/freedesktop/login1"),
			dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		}).
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case "org.freedesktop.login1.Manager.PrepareForSleep":
				if assertTruthiness(s.Body[0]) {
					sensorCh <- newPowerState("Suspended")
				} else {
					sensorCh <- newPowerState("Powered On")
				}
			case "org.freedesktop.login1.Manager.PrepareForShutdown":
				if assertTruthiness(s.Body[0]) {
					sensorCh <- newPowerState("Powered Off")
				} else {
					sensorCh <- newPowerState("Powered On")
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Failed to create user D-Bus watch. Will not track power state.")
		close(sensorCh)
		return sensorCh
	}

	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped power state sensor.")
	}()
	return sensorCh
}

func newPowerState(state string) *powerStateSensor {
	return &powerStateSensor{
		linuxSensor: linuxSensor{
			sensorType: powerState,
			value:      state,
			source:     srcDbus,
			diagnostic: true,
		},
	}
}

func assertTruthiness(v any) bool {
	if isTrue, ok := v.(bool); ok && isTrue {
		return true
	}
	return false
}
