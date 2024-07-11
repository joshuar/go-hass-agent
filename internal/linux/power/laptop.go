// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dockedProp        = managerInterface + ".Docked"
	lidClosedProp     = managerInterface + ".LidClosed"
	externalPowerProp = managerInterface + ".OnExternalPower"

	laptopWorkerID = "laptop_sensors"
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

type laptopWorker struct {
	logger *slog.Logger
	bus    *dbusx.Bus
}

//nolint:cyclop,gocognit
func (w *laptopWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	// If we can't get a session path, we can't run.
	sessionPath, err := w.bus.GetSessionPath(ctx)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not create laptop worker: %w", err)
	}

	triggerCh, err := w.bus.WatchBus(ctx, &dbusx.Watch{
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
				return
			case event := <-triggerCh:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					w.logger.Warn("Received unknown event from D-Bus.", "error", err.Error())

					continue
				}

				for prop, value := range props.Changed {
					if slices.Contains(laptopPropList, prop) {
						if state, err := dbusx.VariantToValue[bool](value); err != nil {
							w.logger.Warn("Could not parse property value.", "property", prop, "error", err.Error())
						} else {
							sensorCh <- newLaptopEvent(prop, state)
						}
					}
				}
			}
		}
	}()

	// Send an initial update.
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			w.logger.Warn("Could not retrieve laptop properties from D-Bus.", "error", err.Error())
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
		state, err := dbusx.GetProp[bool](ctx, w.bus, loginBasePath, loginBaseInterface, prop)
		if err != nil {
			w.logger.Debug("Could not retrieve property", "property", filepath.Ext(prop), "error", err.Error())

			continue
		}

		sensors = append(sensors, newLaptopEvent(prop, state))
	}

	return sensors, nil
}

func NewLaptopWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	// Don't run this worker if we are not running on a laptop.
	chassis, _ := device.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	if chassis != "laptop" {
		return nil, fmt.Errorf("unable to monitor laptop sensors: %w", device.ErrUnsupportedHardware)
	}

	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor laptop sensors: %w", err)
	}

	return &linux.SensorWorker{
			Value: &laptopWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", laptopWorkerID)),
				bus:    bus,
			},
			WorkerID: laptopWorkerID,
		},
		nil
}
