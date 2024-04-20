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
	powerProfilesDBusPath = "/net/hadess/PowerProfiles"
	powerProfilesDBusDest = "net.hadess.PowerProfiles"
)

type powerSensor struct {
	linux.Sensor
}

func newPowerSensor(t linux.SensorTypeValue, v dbus.Variant) *powerSensor {
	s := &powerSensor{}
	s.Value = dbusx.VariantToValue[string](v)
	s.SensorTypeValue = t
	s.IconString = "mdi:flash"
	s.SensorSrc = linux.DataSrcDbus
	s.IsDiagnostic = true
	return s
}

func ProfileUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(powerProfilesDBusPath).
		Destination(powerProfilesDBusDest)
	activePowerProfile, err := dbusx.GetProp[dbus.Variant](req, powerProfilesDBusDest+".ActiveProfile")
	if err != nil {
		log.Debug().Err(err).Msg("Cannot retrieve a power profile from D-Bus. Will not run power sensor.")
		close(sensorCh)
		return sensorCh
	}

	go func() {
		sensorCh <- newPowerSensor(linux.SensorPowerProfile, activePowerProfile)
	}()

	err = dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchInterface(powerProfilesDBusDest),
			dbus.WithMatchObjectPath(powerProfilesDBusPath),
			dbus.WithMatchMember("ActiveProfile"),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Name != dbusx.PropChangedSignal || s.Path != powerProfilesDBusPath {
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
					sensorCh <- newPowerSensor(linux.SensorPowerProfile, propValue)
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
