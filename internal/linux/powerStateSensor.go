// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/rs/zerolog/log"
)

type powerStateSensor struct {
	linuxSensor
}

func (s *powerStateSensor) Icon() string {
	switch s.value.(string) {
	case "Suspended":
		return "mdi:power-sleep"
	case "Powered Off":
		return "mdi:power-off"
	}
	return "mdi:power-on"
}

func PowerStateUpdater(ctx context.Context, tracker device.SensorTracker) {
	sensorCh := make(chan interface{})
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	go func() {
		for sensor := range sensorCh {
			if err := tracker.UpdateSensors(ctx, sensor); err != nil {
				log.Error().Err(err).Msg("Could not update power state sensor.")
			}
		}
	}()

	sensorCh <- newPowerState("Powered On")

	err := NewBusRequest(ctx, SystemBus).
		Path("/org/freedesktop/login1").
		Match([]dbus.MatchOption{
			dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
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
			Msg("Failed to create user D-Bus watch. Will not track powered off state.")
	}
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

func assertTruthiness(v interface{}) bool {
	if isTrue, ok := v.(bool); ok && isTrue {
		return true
	}
	return false
}
