// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package power

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dockedProp        = managerInterface + ".Docked"
	lidClosedProp     = managerInterface + ".LidClosed"
	externalPowerProp = managerInterface + ".OnExternalPower"

	laptopWorkerID     = "laptop_sensors"
	laptopWorkerPrefID = sensorsPrefPrefix + "laptop"
)

var laptopPropList = []string{dockedProp, lidClosedProp, externalPowerProp}

var ErrInitLaptopWorker = errors.New("could not init laptop worker")

func newLaptopEvent(prop string, state bool) sensor.Entity {
	var (
		name, icon  string
		deviceClass types.DeviceClass
	)

	switch prop {
	case dockedProp:
		name = "Docked State"

		if state {
			icon = "mdi:desktop-tower-monitor"
		} else {
			icon = "mdi:laptop"
		}

		deviceClass = types.BinarySensorDeviceClassConnectivity
	case lidClosedProp:
		name = "Lid Closed"

		if state {
			icon = "mdi:laptop"
		} else {
			icon = "mdi:laptop-off"
		}

		deviceClass = types.BinarySensorDeviceClassOpening
		state = !state // Invert state for BinarySensorDeviceClassOpening: On means open, Off means closed.
	case externalPowerProp:
		name = "External Power Connected"

		if state {
			icon = "mdi:power-plug"
		} else {
			icon = "mdi:battery"
		}

		deviceClass = types.BinarySensorDeviceClassPower
	}

	return sensor.NewSensor(
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(deviceClass),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithValue(state),
			sensor.WithDataSourceAttribute(linux.DataSrcDbus),
		),
	)
}

type laptopWorker struct {
	triggerCh  chan dbusx.Trigger
	properties map[string]*dbusx.Property[bool]
	prefs      *preferences.CommonWorkerPrefs
}

func (w *laptopWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

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

func (w *laptopWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	sensors := make([]sensor.Entity, 0, len(laptopPropList))

	// For each property, get its current state as a sensor.
	for name, prop := range w.properties {
		state, err := prop.Get()
		if err != nil {
			slog.With(slog.String("worker", laptopWorkerID)).
				Debug("Could not retrieve property",
					slog.String("property", name),
					slog.Any("error", err))
		} else {
			sensors = append(sensors, newLaptopEvent(name, state))
		}
	}

	return sensors, nil
}

func (w *laptopWorker) PreferencesID() string {
	return laptopWorkerPrefID
}

func (w *laptopWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

//nolint:nilnil
func NewLaptopWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventSensorWorker(laptopWorkerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, errors.Join(ErrInitLaptopWorker, linux.ErrNoSystemBus)
	}

	// If we can't get a session path, we can't run.
	sessionPath, ok := linux.CtxGetSessionPath(ctx)
	if !ok {
		return worker, linux.ErrNoSessionPath
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(sessionPath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers("PropertiesChanged"),
	).Start(ctx, bus)
	if err != nil {
		return worker, errors.Join(ErrInitLaptopWorker,
			fmt.Errorf("unable to create D-Bus watch for laptop property updates: %w", err))
	}

	properties := make(map[string]*dbusx.Property[bool])
	for _, name := range laptopPropList {
		properties[name] = dbusx.NewProperty[bool](bus, loginBasePath, loginBaseInterface, name)
	}

	eventWorker := &laptopWorker{
		triggerCh:  triggerCh,
		properties: properties,
	}

	eventWorker.prefs, err = preferences.LoadWorker(eventWorker)
	if err != nil {
		return nil, errors.Join(ErrInitLaptopWorker, err)
	}

	if eventWorker.prefs.IsDisabled() {
		return nil, nil
	}

	worker.EventSensorType = eventWorker

	return worker, nil
}

func sendChangedProps(props map[string]dbus.Variant, sensorCh chan sensor.Entity) {
	for prop, value := range props {
		if slices.Contains(laptopPropList, prop) {
			if state, err := dbusx.VariantToValue[bool](value); err != nil {
				slog.With(slog.String("worker", laptopWorkerID)).
					Debug("Could not parse property value.",
						slog.String("property", prop),
						slog.Any("error", err))
			} else {
				sensorCh <- newLaptopEvent(prop, state)
			}
		}
	}
}
