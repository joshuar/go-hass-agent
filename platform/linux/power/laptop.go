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
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*laptopWorker)(nil)

var laptopPropList = []string{dockedProp, lidClosedProp, externalPowerProp}

var (
	ErrNewLaptopSensor  = errors.New("could not create laptop sensor")
	ErrInitLaptopWorker = errors.New("could not init laptop worker")
)

const (
	dockedProp        = managerInterface + ".Docked"
	lidClosedProp     = managerInterface + ".LidClosed"
	externalPowerProp = managerInterface + ".OnExternalPower"

	laptopWorkerID     = "laptop_sensors"
	laptopWorkerDesc   = "Laptop sensors"
	laptopWorkerPrefID = sensorsPrefPrefix + "laptop"
)

type laptopWorker struct {
	bus         *dbusx.Bus
	sessionPath string
	properties  map[string]*dbusx.Property[bool]
	prefs       *workers.CommonWorkerPrefs
	*models.WorkerMetadata
}

func newLaptopEvent(ctx context.Context, prop string, state bool) models.Entity {
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
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
	)
}

func (w *laptopWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(w.sessionPath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers("PropertiesChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, errors.Join(ErrInitLaptopWorker,
			fmt.Errorf("unable to create D-Bus watch for laptop property updates: %w", err))
	}
	sensorCh := make(chan models.Entity)

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
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
		sensors, err := w.generateSensors(ctx)
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

func (w *laptopWorker) generateSensors(ctx context.Context) ([]models.Entity, error) {
	var warnings error

	sensors := make([]models.Entity, 0, len(laptopPropList))

	// For each property, get its current state as a sensor.
	for name, prop := range w.properties {
		state, err := prop.Get()
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not retrieve property from D-Bus: %w", err))
		} else {
			sensors = append(sensors, newLaptopEvent(ctx, name, state))
		}
	}

	return sensors, warnings
}

func (w *laptopWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func NewLaptopWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitLaptopWorker, linux.ErrNoSystemBus)
	}
	// If we can't get a session path, we can't run.
	sessionPath, ok := linux.CtxGetSessionPath(ctx)
	if !ok {
		return nil, linux.ErrNoSessionPath
	}
	properties := make(map[string]*dbusx.Property[bool])
	for _, name := range laptopPropList {
		properties[name] = dbusx.NewProperty[bool](bus, loginBasePath, loginBaseInterface, name)
	}

	worker := &laptopWorker{
		WorkerMetadata: models.SetWorkerMetadata(laptopWorkerID, laptopWorkerDesc),
		bus:            bus,
		sessionPath:    sessionPath,
		properties:     properties,
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(laptopWorkerPrefID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitLaptopWorker, err)
	}

	return worker, nil
}

func sendChangedProps(ctx context.Context, props map[string]dbus.Variant, sensorCh chan models.Entity) {
	for prop, value := range props {
		if slices.Contains(laptopPropList, prop) {
			if state, err := dbusx.VariantToValue[bool](value); err != nil {
				slogctx.FromCtx(ctx).Warn("Could not parse laptop D-Bus property.", slog.Any("error", err))
			} else {
				sensorCh <- newLaptopEvent(ctx, prop, state)
			}
		}
	}
}
