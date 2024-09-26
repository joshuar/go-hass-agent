// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
//go:generate stringer -type=batterySensor -output battery_generated.go -linecomment
package battery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	battType       batterySensor = iota // Battery Type
	battPercentage                      // Battery Level
	battTemp                            // Battery Temperature
	battVoltage                         // Battery Voltage
	battEnergy                          // Battery Energy
	battEnergyRate                      // Battery Power
	battState                           // Battery State
	battNativePath                      // Battery Path
	battLevel                           // Battery Level
	battModel                           // Battery Model

	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusDeviceDest   = upowerDBusDest + ".Device"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	deviceAddedSignal   = "DeviceAdded"
	deviceRemovedSignal = "DeviceRemoved"

	batteryIcon = "mdi:battery"

	workerID = "battery_sensors"
)

type batterySensor int

var ErrInvalidBattery = errors.New("invalid battery")

// dBusSensorToProps is a map of battery sensors to their D-Bus properties.
var dBusSensorToProps = map[batterySensor]string{
	battType:       upowerDBusDeviceDest + ".Type",
	battPercentage: upowerDBusDeviceDest + ".Percentage",
	battTemp:       upowerDBusDeviceDest + ".Temperature",
	battVoltage:    upowerDBusDeviceDest + ".Voltage",
	battEnergy:     upowerDBusDeviceDest + ".Energy",
	battEnergyRate: upowerDBusDeviceDest + ".EnergyRate",
	battState:      upowerDBusDeviceDest + ".State",
	battNativePath: upowerDBusDeviceDest + ".NativePath",
	battLevel:      upowerDBusDeviceDest + ".BatteryLevel",
	battModel:      upowerDBusDeviceDest + ".Model",
}

// dBusPropToSensor provides a map for to convert D-Bus properties to sensors.
var dBusPropToSensor = map[string]batterySensor{
	"Energy":       battEnergy,
	"EnergyRate":   battEnergyRate,
	"Voltage":      battVoltage,
	"Percentage":   battPercentage,
	"Temperatute":  battTemp,
	"State":        battState,
	"BatteryLevel": battLevel,
}

type upowerBattery struct {
	logger   *slog.Logger
	bus      *dbusx.Bus
	id       string
	model    string
	dBusPath dbus.ObjectPath
	sensors  []batterySensor
	battType batteryType
}

// getProp retrieves the property from D-Bus that matches the given battery sensor type.
func (b *upowerBattery) getProp(t batterySensor) (dbus.Variant, error) {
	value, err := dbusx.NewProperty[dbus.Variant](b.bus, string(b.dBusPath), upowerDBusDest, dBusSensorToProps[t]).Get()
	if err != nil {
		return dbus.Variant{}, fmt.Errorf("could not retrieve battery property %s: %w", t.String(), err)
	}

	return value, nil
}

// getSensors retrieves the sensors passed in for a given battery.
func (b *upowerBattery) getSensors(sensors ...batterySensor) chan sensor.Entity {
	sensorCh := make(chan sensor.Entity, len(sensors))
	defer close(sensorCh)

	for _, batterySensor := range sensors {
		value, err := b.getProp(batterySensor)
		if err != nil {
			b.logger.Warn("Could not retrieve battery sensor.",
				slog.String("sensor", batterySensor.String()),
				slog.Any("error", err))

			continue
		}
		sensorCh <- newBatterySensor(b, batterySensor, value)
	}

	return sensorCh
}

// newBattery creates a battery object that will have a number of properties to
// be treated as sensors in Home Assistant.
func newBattery(bus *dbusx.Bus, logger *slog.Logger, path dbus.ObjectPath) (*upowerBattery, error) {
	battery := &upowerBattery{
		dBusPath: path,
		bus:      bus,
	}

	var (
		variant dbus.Variant
		err     error
	)

	// Get the battery type. Depending on the value, additional sensors will be added.
	variant, err = battery.getProp(battType)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}
	// Store the battery type.
	battery.battType, err = dbusx.VariantToValue[batteryType](variant)
	if err != nil {
		return nil, fmt.Errorf("could not determine battery type: %w", err)
	}

	// Use the native path D-Bus property for the battery id.
	variant, err = battery.getProp(battNativePath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}
	// Store the battery id/name.
	battery.id, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve battery path in D-Bus: %w", err)
	}

	// Set up a logger for the battery with some battery-specific default
	// attributes.
	battery.logger = logger.With(
		slog.Group("battery_info",
			slog.String("name", battery.id),
			slog.String("dbus_path", string(battery.dBusPath)),
		),
	)

	// Get the battery model.
	variant, err = battery.getProp(battModel)
	if err != nil {
		battery.logger.Warn("Could not determine battery model.")
	}
	// Store the battery model.
	battery.model, err = dbusx.VariantToValue[string](variant)
	if err != nil {
		battery.logger.Warn("Could not determine battery model.")
	}

	// At a minimum, monitor the battery type and the charging state.
	battery.sensors = append(battery.sensors, battState)

	if battery.battType == batteryTypeBattery {
		// Battery has charge percentage, temp and charging rate sensors
		battery.sensors = append(battery.sensors, battPercentage, battTemp, battEnergyRate)
	} else {
		// Battery has a textual level sensor
		battery.sensors = append(battery.sensors, battLevel)
	}

	return battery, nil
}

// monitorBattery will monitor a battery device for any property changes and
// send these as sensors.
func monitorBattery(ctx context.Context, battery *upowerBattery) <-chan sensor.Entity {
	sensorCh := make(chan sensor.Entity)
	// Create a DBus signal match to watch for property changes for this
	// battery.
	events, err := dbusx.NewWatch(
		dbusx.MatchPath(string(battery.dBusPath)),
		dbusx.MatchPropChanged(),
	).Start(ctx, battery.bus)
	if err != nil {
		battery.logger.Debug("Failed to create D-Bus watch for battery property changes.", slog.Any("error", err))
		close(sensorCh)

		return sensorCh
	}

	go func() {
		battery.logger.Debug("Monitoring battery.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				battery.logger.Debug("Stopped monitoring battery.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					battery.logger.Warn("Received a battery property change event that could not be understood.", slog.Any("error", err))

					continue
				}

				for prop, value := range props.Changed {
					if s, ok := dBusPropToSensor[prop]; ok {
						sensorCh <- newBatterySensor(battery, s, value)
					}
				}
			}
		}
	}()

	return sensorCh
}
