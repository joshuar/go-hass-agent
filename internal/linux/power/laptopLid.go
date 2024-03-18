// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type laptopLidSensor struct {
	linux.Sensor
}

func (s *laptopLidSensor) Icon() string {
	state, ok := s.Value.(bool)
	if !ok {
		return "mdi:lock-alert"
	}
	if state {
		return "mdi:laptop"
	}
	return "mdi:laptop-off"
}

func newLaptopLidEvent(v bool) *laptopLidSensor {
	return &laptopLidSensor{
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorLaptopLid,
			IsBinary:        true,
			SensorSrc:       linux.DataSrcDbus,
			Value:           v,
		},
	}
}

func LaptopLidUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath("/org/freedesktop/login1/session"),
			dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), "/org/freedesktop/login1") || len(s.Body) <= 1 {
				log.Trace().Str("runner", "power").Msg("Not my signal or empty signal body.")
				return
			}

			if s.Name == dbusx.PropChangedSignal {
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					log.Trace().Str("runner", "power").
						Str("signal", s.Name).Interface("body", s.Body).
						Msg("Unexpected signal body")
					return
				}
				if v, ok := props["LidClosed"]; ok {
					sensorCh <- newLaptopLidEvent(!dbusx.VariantToValue[bool](v))
				}
			}

		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not poll D-Bus for laptopLid. LaptopLid sensor will not run.")
		close(sensorCh)
		return sensorCh
	}
	log.Trace().Msg("Started laptopLid sensor.")
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Trace().Msg("Stopped laptopLid sensor.")
	}()
	return sensorCh
}
