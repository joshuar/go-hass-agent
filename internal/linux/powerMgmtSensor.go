// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

type powerMgmtSensor struct {
	linuxSensor
}

func (s *powerMgmtSensor) SensorType() sensor.SensorType {
	return sensor.TypeBinary
}

func (s *powerMgmtSensor) Icon() string {
	if isTrue, ok := s.value.(bool); ok && isTrue {
		switch s.sensorType {
		case isSuspended:
			return "mdi:power-sleep"
		case isShutdown:
			return "mdi:power-off"
		}
	}
	return "mdi:power-on"
}

func PowerMgmtUpdater(ctx context.Context, tracker device.SensorTracker) {
	if err := tracker.UpdateSensors(ctx, newPowerMgmtSensor(isShutdown, false), newPowerMgmtSensor(isSuspended, false)); err != nil {
		log.Error().Err(err).Msg("Could not update power management sensors.")
	}

	r := NewBusRequest(ctx, SystemBus).
		Path("/org/freedesktop/login1").
		Match([]dbus.MatchOption{
			dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		})

	err := NewBusRequest(ctx, SystemBus).
		Path("/org/freedesktop/login1").
		Match([]dbus.MatchOption{
			dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		}).Event("org.freedesktop.login1.Manager.PrepareForSleep").
		Handler(func(s *dbus.Signal) {
			if s.Name != "org.freedesktop.login1.Manager.PrepareForSleep" {
				return
			}
			if err := tracker.UpdateSensors(ctx, newPowerMgmtSensor(isSuspended, s.Body[0].(bool))); err != nil {
				log.Error().Err(err).Msg("Could not update suspended sensor.")
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Failed to create user D-Bus watch. Will not track suspend state.")
	}
	err = r.Event("org.freedesktop.login1.Manager.PrepareForShutdown").
		Handler(func(s *dbus.Signal) {
			if s.Name != "org.freedesktop.login1.Manager.PrepareForShutdown" {
				return
			}
			if err = tracker.UpdateSensors(ctx, newPowerMgmtSensor(isShutdown, s.Body[0].(bool))); err != nil {
				log.Error().Err(err).Msg("Could not update powered off sensor.")
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Failed to create user D-Bus watch. Will not track powered off state.")
	}
}

func newPowerMgmtSensor(t sensorType, v interface{}) *powerMgmtSensor {
	return &powerMgmtSensor{
		linuxSensor: linuxSensor{
			value:      v,
			sensorType: t,
			source:     srcDbus,
			diagnostic: true,
		},
	}
}
