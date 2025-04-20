// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package battery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusDeviceDest   = upowerDBusDest + ".Device"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	deviceAddedSignal   = "DeviceAdded"
	deviceRemovedSignal = "DeviceRemoved"

	batteryIcon = "mdi:battery"
)

var ErrInvalidBattery = errors.New("invalid battery")

// dBusSensorToProps is a map of battery sensors to their D-Bus properties.
var dBusSensorToProps = map[sensorType]string{
	typeDesc:       upowerDBusDeviceDest + ".Type",
	typePercentage: upowerDBusDeviceDest + ".Percentage",
	typeTemp:       upowerDBusDeviceDest + ".Temperature",
	typeVoltage:    upowerDBusDeviceDest + ".Voltage",
	typeEnergy:     upowerDBusDeviceDest + ".Energy",
	typeEnergyRate: upowerDBusDeviceDest + ".EnergyRate",
	typeState:      upowerDBusDeviceDest + ".State",
	typeNativePath: upowerDBusDeviceDest + ".NativePath",
	typeLevel:      upowerDBusDeviceDest + ".BatteryLevel",
	typeModel:      upowerDBusDeviceDest + ".Model",
}

// dBusPropToSensor provides a map for to convert D-Bus properties to sensors.
var dBusPropToSensor = map[string]sensorType{
	"Energy":       typeEnergy,
	"EnergyRate":   typeEnergyRate,
	"Voltage":      typeVoltage,
	"Percentage":   typePercentage,
	"Temperatute":  typeTemp,
	"State":        typeState,
	"BatteryLevel": typeLevel,
}

// upowerBattery contains the data to represent a battery as derived from the
// upower D-Bus.
type upowerBattery struct {
	bus      *dbusx.Bus
	id       string
	model    string
	dBusPath dbus.ObjectPath
	sensors  []sensorType
	battType typeDescription
}

// getProp retrieves the property from D-Bus that matches the given battery sensor type.
func (b *upowerBattery) getProp(t sensorType) (dbus.Variant, error) {
	value, err := dbusx.NewProperty[dbus.Variant](b.bus, string(b.dBusPath), upowerDBusDest, dBusSensorToProps[t]).Get()
	if err != nil {
		return dbus.Variant{}, fmt.Errorf("could not retrieve battery property %s: %w", t.String(), err)
	}

	return value, nil
}

// getSensors retrieves the sensors passed in for a given battery.
func (b *upowerBattery) getSensors(ctx context.Context, sensors ...sensorType) chan models.Entity {
	sensorCh := make(chan models.Entity, len(sensors))
	defer close(sensorCh)

	for _, batterySensor := range sensors {
		value, err := b.getProp(batterySensor)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not retrieve battery models.",
				slog.String("sensor", batterySensor.String()),
				slog.Any("error", err))

			continue
		}

		entity, err := newBatterySensor(ctx, b, batterySensor, value)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not generate battery sensor.",
				slog.String("sensor", batterySensor.String()),
				slog.Any("error", err))

			continue
		}

		sensorCh <- entity
	}

	return sensorCh
}

// newBattery creates a battery object that will have a number of properties to
// be treated as sensors in Home Assistant.
func newBattery(ctx context.Context, bus *dbusx.Bus, path dbus.ObjectPath) (*upowerBattery, error) {
	battery := &upowerBattery{
		dBusPath: path,
		bus:      bus,
	}

	var (
		variant dbus.Variant
		err     error
	)

	// Get the battery type. Depending on the value, additional sensors will be added.
	variant, err = battery.getProp(typeDesc)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}
	// Store the battery type.
	battery.battType, err = dbusx.VariantToValue[typeDescription](variant)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}

	// Use the native path D-Bus property for the battery id.
	variant, err = battery.getProp(typeNativePath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}
	// Store the battery id/name.
	battery.id, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}

	// Get the battery model.
	variant, err = battery.getProp(typeModel)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not determine battery model.")
	}
	// Store the battery model.
	battery.model, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not determine battery model.")
	}

	// At a minimum, monitor the battery type and the charging state.
	battery.sensors = append(battery.sensors, typeState)

	if battery.battType == batteryType {
		// Battery has charge percentage, temp and charging rate sensors
		battery.sensors = append(battery.sensors, typePercentage, typeTemp, typeEnergyRate, typeVoltage, typeEnergy)
	} else {
		// Battery has a textual level sensor
		battery.sensors = append(battery.sensors, typeLevel)
	}

	return battery, nil
}

// monitorBattery will monitor a battery device for any property changes and
// send these as sensors.
//
//nolint:gocognit
func monitorBattery(ctx context.Context, battery *upowerBattery) <-chan models.Entity {
	sensorCh := make(chan models.Entity)
	// Create a DBus signal match to watch for property changes for this
	// battery.
	events, err := dbusx.NewWatch(
		dbusx.MatchPath(string(battery.dBusPath)),
		dbusx.MatchPropChanged(),
	).Start(ctx, battery.bus)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Failed to create D-Bus watch for battery property changes.", slog.Any("error", err))
		close(sensorCh)

		return sensorCh
	}

	go func() {
		slogctx.FromCtx(ctx).Debug("Monitoring battery.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Stopped monitoring battery.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Received a battery property change event that could not be understood.", slog.Any("error", err))

					continue
				}

				for prop, value := range props.Changed {
					if s, ok := dBusPropToSensor[prop]; ok {
						entity, err := newBatterySensor(ctx, battery, s, value)
						if err != nil {
							slogctx.FromCtx(ctx).Warn("Could not generate battery sensor.",
								slog.String("sensor", s.String()),
								slog.Any("error", err))

							continue
						}

						sensorCh <- entity
					}
				}
			}
		}
	}()

	return sensorCh
}
