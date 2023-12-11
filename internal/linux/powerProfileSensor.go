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

const (
	powerProfilesDBusPath = "/net/hadess/PowerProfiles"
	powerProfilesDBusDest = "net.hadess.PowerProfiles"
)

type powerSensor struct {
	linuxSensor
}

func newPowerSensor(t sensorType, v dbus.Variant) *powerSensor {
	s := &powerSensor{}
	s.value = dbushelpers.VariantToValue[string](v)
	s.sensorType = t
	s.icon = "mdi:flash"
	s.source = srcDbus
	s.isDiagnostic = true
	return s
}

func PowerProfileUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	activePowerProfile, err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(powerProfilesDBusPath).
		Destination(powerProfilesDBusDest).
		GetProp(powerProfilesDBusDest + ".ActiveProfile")
	if err != nil {
		log.Debug().Err(err).Msg("Cannot retrieve a power profile from D-Bus. Will not run power sensor.")
		close(sensorCh)
		return sensorCh
	}

	sensorCh <- newPowerSensor(powerProfile, activePowerProfile)

	err = dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchInterface(powerProfilesDBusDest),
			dbus.WithMatchObjectPath(powerProfilesDBusPath),
			dbus.WithMatchMember("ActiveProfile"),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Name != dbushelpers.PropChangedSignal || s.Path != powerProfilesDBusPath {
				return
			}
			if len(s.Body) <= 1 {
				log.Debug().Caller().Interface("body", s.Body).Msg("Unexpected body length.")
				return
			}
			updatedProps, ok := s.Body[1].(map[string]dbus.Variant)
			if !ok {
				log.Debug().Caller().
					Str("signal", s.Name).Interface("body", s.Body).
					Msg("Unexpected signal body")
				return
			}
			for propName, propValue := range updatedProps {
				if propName == "ActiveProfile" {
					sensorCh <- newPowerSensor(powerProfile, propValue)
				} else {
					log.Debug().Msgf("Unhandled property %v changed to %v", propName, propValue)
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create power state D-Bus watch.")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg(("Stopped power profile sensor."))
	}()
	return sensorCh
}
