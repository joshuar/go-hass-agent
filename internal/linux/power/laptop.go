// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

import (
	"context"
	"path/filepath"
	"slices"

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

var laptopPropList = []string{dockedProp, lidClosedProp, externalPowerProp}

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

	for _, prop := range laptopPropList {
		go func(p string) {
			sendLaptopPropState(ctx, p, sensorCh)
		}(prop)
	}

	sessionPath := dbusx.GetSessionPath(ctx)

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{dbusx.PropChangedSignal},
		Interface: managerInterface,
		Path:      string(sessionPath),
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create laptop sensors D-Bus watch.")
		close(sensorCh)
		return sensorCh
	}

	go func() {
		defer close(sensorCh)
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped laptop state sensor.")
				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					log.Warn().Err(err).Msg("Did not understand received trigger.")
					continue
				}
				for prop, value := range props.Changed {
					if slices.Contains(laptopPropList, prop) {
						sensorCh <- newLaptopEvent(prop, dbusx.VariantToValue[bool](value))
					}
				}
			}
		}
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
