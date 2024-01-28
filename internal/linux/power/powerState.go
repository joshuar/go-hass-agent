// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown
)

type powerSignal int

type powerStateSensor struct {
	linux.Sensor
	signal powerSignal
}

func (s *powerStateSensor) State() any {
	b, ok := s.Value.(bool)
	if !ok {
		return sensor.StateUnknown
	}
	if b {
		switch s.signal {
		case suspend:
			return "Suspended"
		case shutdown:
			return "Powered Off"
		}
	}
	return "Powered On"
}

func (s *powerStateSensor) Icon() string {
	str, ok := s.State().(string)
	if !ok {
		str = ""
	}
	switch str {
	case "Suspended":
		return "mdi:power-sleep"
	case "Powered Off":
		return "mdi:power-off"
	default:
		return "mdi:power-on"
	}
}

func newPowerState(s powerSignal, v any) *powerStateSensor {
	return &powerStateSensor{
		signal: s,
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorPowerState,
			Value:           v,
			SensorSrc:       linux.DataSrcDbus,
			IsDiagnostic:    true,
		},
	}
}

func PowerStateUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)

	sensorCh <- newPowerState(shutdown, false)

	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath("/org/freedesktop/login1"),
			dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		}).
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case "org.freedesktop.login1.Manager.PrepareForSleep":
				sensorCh <- newPowerState(suspend, s.Body[0])
			case "org.freedesktop.login1.Manager.PrepareForShutdown":
				sensorCh <- newPowerState(shutdown, s.Body[0])
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
