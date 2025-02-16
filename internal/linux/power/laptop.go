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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
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

func newLaptopEvent(ctx context.Context, prop string, state bool) (models.Entity, error) {
	var (
		name, icon  string
		deviceClass class.SensorDeviceClass
	)

	switch prop {
	case dockedProp:
		name = "Docked State"

		if state {
			icon = "mdi:desktop-tower-monitor"
		} else {
			icon = "mdi:laptop"
		}

		deviceClass = class.BinaryClassConnectivity
	case lidClosedProp:
		name = "Lid Closed"

		if state {
			icon = "mdi:laptop"
		} else {
			icon = "mdi:laptop-off"
		}

		deviceClass = class.BinaryClassOpening
		state = !state // Invert state for BinarySensorDeviceClassOpening: On means open, Off means closed.
	case externalPowerProp:
		name = "External Power Connected"

		if state {
			icon = "mdi:power-plug"
		} else {
			icon = "mdi:battery"
		}

		deviceClass = class.BinaryClassPower
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(deviceClass),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(state),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
	)
}

type laptopWorker struct {
	triggerCh  chan dbusx.Trigger
	properties map[string]*dbusx.Property[bool]
	prefs      *preferences.CommonWorkerPrefs
}

func (w *laptopWorker) Events(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

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
					sendChangedProps(ctx, props.Changed, sensorCh)
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

func (w *laptopWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	var warnings error

	sensors := make([]models.Entity, 0, len(laptopPropList))

	// For each property, get its current state as a sensor.
	for name, prop := range w.properties {
		state, err := prop.Get()
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not retrieve property from D-Bus: %w", err))
		} else {
			entity, err := newLaptopEvent(ctx, name, state)
			if err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("could not generate laptop sensor: %w", err))
			} else {
				sensors = append(sensors, entity)
			}
		}
	}

	return sensors, warnings
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

func sendChangedProps(ctx context.Context, props map[string]dbus.Variant, sensorCh chan models.Entity) {
	for prop, value := range props {
		if slices.Contains(laptopPropList, prop) {
			if state, err := dbusx.VariantToValue[bool](value); err != nil {
				logging.FromContext(ctx).Warn("Could not parse laptop D-Bus property.", slog.Any("error", err))
			} else {
				if entity, err := newLaptopEvent(ctx, prop, state); err != nil {
					logging.FromContext(ctx).Warn("could not send laptop sensor.", slog.Any("error", err))
				} else {
					sensorCh <- entity
				}
			}
		}
	}
}
