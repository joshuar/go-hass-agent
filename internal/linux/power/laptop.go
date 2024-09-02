// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

func newLaptopEvent(prop string, state bool) *laptopSensor {
	sensorEvent := &laptopSensor{
		prop: prop,
		Sensor: linux.Sensor{
			IsBinary:     true,
			IsDiagnostic: true,
			DataSource:   linux.DataSrcDbus,
			Value:        state,
		},
	}

	switch prop {
	case dockedProp:
		sensorEvent.DisplayName = "Docked State"
	case lidClosedProp:
		sensorEvent.DisplayName = "Lid Closed"
	case externalPowerProp:
		sensorEvent.DisplayName = "External Power Connected"
	}

	return sensorEvent
}

type laptopWorker struct {
	triggerCh  chan dbusx.Trigger
	properties map[string]*dbusx.Property[bool]
}

func (w *laptopWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					slog.With(slog.String("worker", laptopWorkerID)).
						Debug("Received unknown event from D-Bus.", slog.Any("error", err))
				} else {
					sendChangedProps(props.Changed, sensorCh)
				}
			}
		}
	}()

	// Send an initial update.
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			slog.With(slog.String("worker", laptopWorkerID)).
				Debug("Could not retrieve laptop properties from D-Bus.", slog.Any("error", err))
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}()

	return sensorCh, nil
}

func (w *laptopWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, len(laptopPropList))

	// For each property, get its current state as a sensor.
	for name, prop := range w.properties {
		state, err := prop.Get()
		if err != nil {
			slog.With(slog.String("worker", laptopWorkerID)).
				Debug("Could not retrieve property", slog.String("property", name), slog.Any("error", err))
		} else {
			sensors = append(sensors, newLaptopEvent(name, state))
		}
	}

	return sensors, nil
}

func NewLaptopWorker(ctx context.Context) (*linux.SensorWorker, error) {
	// Don't run this worker if we are not running on a laptop.
	chassis, _ := device.Chassis() //nolint:errcheck // error is same as any value other than wanted value.
	if chassis != "laptop" {
		return nil, fmt.Errorf("unable to monitor laptop sensors: %w", device.ErrUnsupportedHardware)
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	// If we can't get a session path, we can't run.
	sessionPath, ok := linux.CtxGetSessionPath(ctx)
	if !ok {
		return nil, linux.ErrNoSessionPath
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(sessionPath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers("PropertiesChanged"),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to create D-Bus watch for laptop property updates: %w", err)
	}

	worker := &laptopWorker{
		properties: make(map[string]*dbusx.Property[bool]),
		triggerCh:  triggerCh,
	}

	// Generate the list of laptop properties to track.
	for _, name := range laptopPropList {
		worker.properties[name] = dbusx.NewProperty[bool](bus, loginBasePath, loginBaseInterface, name)
	}

	return &linux.SensorWorker{
			Value:    worker,
			WorkerID: laptopWorkerID,
		},
		nil
}

func sendChangedProps(props map[string]dbus.Variant, sensorCh chan sensor.Details) {
	for prop, value := range props {
		if slices.Contains(laptopPropList, prop) {
			if state, err := dbusx.VariantToValue[bool](value); err != nil {
				slog.With(slog.String("worker", laptopWorkerID)).
					Debug("Could not parse property value.", slog.String("property", prop), slog.Any("error", err))
			} else {
				sensorCh <- newLaptopEvent(prop, state)
			}
		}
	}
}
