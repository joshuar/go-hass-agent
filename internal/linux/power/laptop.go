// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dockedProp        = managerInterface + ".Docked"
	lidClosedProp     = managerInterface + ".LidClosed"
	externalPowerProp = managerInterface + ".OnExternalPower"
)

type laptopSensor struct {
	prop string
	linux.Sensor
}

func (s *laptopSensor) Icon() string {
	state, ok := s.Value.(bool)
	if !ok {
		return "mdi:alert"
	}
	switch s.prop {
	case dockedProp:
		if state {
			return "mdi:desktop-tower-monitor"
		} else {
			return "mdi:laptop"
		}
	case lidClosedProp:
		if state {
			return "mdi:laptop"
		} else {
			return "mdi:laptop-off"
		}
	case externalPowerProp:
		if state {
			return "mdi:power-plug"
		} else {
			return "mdi:battery"
		}
	}
	return "mdi:help"
}

func newLaptopEvent(prop string, state bool) *laptopSensor {
	sensorEvent := &laptopSensor{
		prop: prop,
		Sensor: linux.Sensor{
			IsBinary:     true,
			IsDiagnostic: true,
			SensorSrc:    linux.DataSrcDbus,
			Value:        state,
		},
	}
	switch prop {
	case dockedProp:
		sensorEvent.SensorTypeValue = linux.SensorDocked
	case lidClosedProp:
		sensorEvent.SensorTypeValue = linux.SensorLidClosed
	case externalPowerProp:
		sensorEvent.SensorTypeValue = linux.SensorExternalPower
	}
	return sensorEvent
}

func LaptopUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	if linux.Chassis() != "laptop" {
		close(sensorCh)
		return sensorCh
	}

	for _, prop := range []string{dockedProp, lidClosedProp, externalPowerProp} {
		go func(p string) {
			sendLaptopPropState(ctx, p, sensorCh)
		}(prop)
	}

	sessionPath := dbusx.GetSessionPath(ctx)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(sessionPath),
			dbus.WithMatchInterface(managerInterface),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), loginBasePath) || len(s.Body) <= 1 {
				return
			}
			if s.Name == dbusx.PropChangedSignal {
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					return
				}
				for k, v := range props {
					switch k {
					case dockedProp:
						go func(state dbus.Variant) {
							sensorCh <- newLaptopEvent(dockedProp, dbusx.VariantToValue[bool](state))
						}(v)
					case lidClosedProp:
						go func(state dbus.Variant) {
							sensorCh <- newLaptopEvent(lidClosedProp, dbusx.VariantToValue[bool](state))
						}(v)
					case externalPowerProp:
						go func(state dbus.Variant) {
							sensorCh <- newLaptopEvent(externalPowerProp, dbusx.VariantToValue[bool](state))
						}(v)
					}
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Could not poll D-Bus for laptop states. Won't report dock/lid/external power state.")
		close(sensorCh)
		return sensorCh
	}
	log.Trace().Msg("Started laptop state sensor.")
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Trace().Msg("Stopped laptop state sensor.")
	}()
	return sensorCh
}

func sendLaptopPropState(ctx context.Context, prop string, outCh chan sensor.Details) {
	var state bool
	var err error
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(loginBasePath).
		Destination(loginBaseInterface)
	if state, err = dbusx.GetProp[bool](req, prop); err != nil {
		log.Debug().Err(err).Str("prop", filepath.Ext(prop)).Msg("Could not retrieve laptop property from D-Bus.")
		return
	}
	outCh <- newLaptopEvent(prop, state)
}
