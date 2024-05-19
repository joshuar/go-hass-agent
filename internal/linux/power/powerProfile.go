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
	powerProfilesPath      = "/net/hadess/PowerProfiles"
	powerProfilesDest      = "net.hadess.PowerProfiles"
	powerProfilesInterface = "org.freedesktop.Upower.PowerProfiles"
	activeProfileProp      = "ActiveProfile"
)

type sensors struct{}

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

func (s *sensors) Sensors(ctx context.Context) []sensor.Details {
	var sensors []sensor.Details
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(powerProfilesPath).
		Destination(powerProfilesDest)
	profile, err := dbusx.GetProp[dbus.Variant](req, powerProfilesDest+"."+activeProfileProp)
	if err != nil {
		log.Debug().Err(err).Msg("Cannot retrieve a power profile from D-Bus.")
		return nil
	}
	return append(sensors, newPowerSensor(linux.SensorPowerProfile, profile))
}

func ProfileUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	s := &sensors{}
	sensors := s.Sensors(ctx)
	if len(sensors) < 1 {
		log.Debug().Msg("No power profile detected. Not monitoring power profiles")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		for _, sensor := range sensors {
			sensorCh <- sensor
		}
	}()

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{dbusx.PropChangedSignal},
		Interface: dbusx.PropInterface,
		Path:      powerProfilesPath,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create power profile D-Bus watch.")
		close(sensorCh)
		return sensorCh
	}
	go func() {
		defer close(sensorCh)
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg(("Stopped power profile sensor."))
				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					log.Warn().Err(err).Msg("Did not understand received trigger.")
					continue
				}
				if profile, profileChanged := props.Changed[activeProfileProp]; profileChanged {
					sensorCh <- newPowerSensor(linux.SensorPowerProfile, profile)
				}
			}
		}
	}()
	return sensorCh
}
