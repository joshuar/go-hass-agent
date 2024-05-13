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
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	suspend powerSignal = iota
	shutdown

	sleepSignal    = managerInterface + ".PrepareForSleep"
	shutdownSignal = managerInterface + ".PrepareForShutdown"
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

func StateUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	go func() {
		sensorCh <- newPowerState(shutdown, false)
	}()

	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(loginBasePath),
			dbus.WithMatchInterface(managerInterface),
		}).
		Handler(func(s *dbus.Signal) {
			switch s.Name {
			case sleepSignal:
				go func() {
					sensorCh <- newPowerState(suspend, s.Body[0])
				}()
			case shutdownSignal:
				go func() {
					sensorCh <- newPowerState(shutdown, s.Body[0])
				}()
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
