// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/rs/zerolog/log"
)

const (
	powerProfilesDBusPath = "/net/hadess/PowerProfiles"
	powerProfilesDBusDest = "net.hadess.PowerProfiles"
)

type powerSensor struct {
	sensorGroup string
	linuxSensor
}

func newPowerSensor(t sensorType, g string, v dbus.Variant) *powerSensor {
	s := &powerSensor{
		sensorGroup: g,
	}
	s.value = strings.Trim(v.String(), "\"")
	s.sensorType = t
	s.icon = "mdi:flash"
	s.source = srcDbus
	s.diagnostic = true
	return s
}

func PowerUpater(ctx context.Context, tracker device.SensorTracker) {
	activePowerProfile, err := NewBusRequest(SystemBus).
		Path(powerProfilesDBusPath).
		Destination(powerProfilesDBusDest).
		GetProp(powerProfilesDBusDest + ".ActiveProfile")
	if err != nil {
		log.Debug().Err(err).Msg("Cannot retrieve a power profile from DBus.")
		return
	}

	tracker.UpdateSensors(ctx, newPowerSensor(powerProfile, powerProfilesDBusPath, activePowerProfile))

	err = NewBusRequest(SystemBus).
		Path(powerProfilesDBusPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(powerProfilesDBusPath),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			updatedProps := s.Body[1].(map[string]dbus.Variant)
			for propName, propValue := range updatedProps {
				if propName == "ActiveProfile" {
					tracker.UpdateSensors(ctx, newPowerSensor(powerProfile, string(s.Path), activePowerProfile))
				} else {
					log.Debug().Msgf("Unhandled property %v changed to %v", propName, propValue)
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create power state DBus watch.")
	}
}
