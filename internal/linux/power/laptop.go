// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
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

//nolint:exhaustruct
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

type laptopWorker struct{}

//nolint:cyclop,exhaustruct
func (w *laptopWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	// If we can't get a session path, we can't run.
	sessionPath, err := dbusx.GetSessionPath(ctx)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not create laptop worker: %w", err)
	}

	triggerCh, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{dbusx.PropChangedSignal},
		Interface: managerInterface,
		Path:      sessionPath,
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for laptop updates: %w", err)
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped laptop state sensor.")

				return
			case event := <-triggerCh:
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

	// Send an initial update.
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Could not get initial sensor updates.")
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}()

	return sensorCh, nil
}

func (w *laptopWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, len(laptopPropList))

	for _, prop := range laptopPropList {
		state, err := dbusx.GetProp[bool](ctx, dbusx.SystemBus, loginBasePath, loginBaseInterface, prop)
		if err != nil {
			log.Debug().Err(err).Str("prop", filepath.Ext(prop)).Msg("Could not retrieve laptop property from D-Bus.")

			continue
		}

		sensors = append(sensors, newLaptopEvent(prop, state))
	}

	return sensors, nil
}

func NewLaptopWorker() (*linux.SensorWorker, error) {
	// Don't run this worker if we are not running on a laptop.
	if linux.Chassis() != "laptop" {
		return nil, fmt.Errorf("will not create laptop sensors: %w", linux.ErrUnsupportedHardware)
	}

	return &linux.SensorWorker{
			WorkerName: "Laptop State Sensors",
			WorkerDesc: "Sensors for laptop lid, dock and external power states.",
			Value:      &laptopWorker{},
		},
		nil
}
