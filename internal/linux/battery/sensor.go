// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package battery

import (
	"fmt"
	"math"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

// newBatterySensor creates a new sensor for Home Assistant from a battery
// property.
func newBatterySensor(battery *upowerBattery, sensorType sensorType, value dbus.Variant) sensor.Entity {
	var (
		name, id, icon, units string
		deviceClass           types.DeviceClass
		stateClass            types.StateClass
	)

	if battery.model == "" {
		name = battery.id + " " + sensorType.String()
	} else {
		name = battery.model + " " + sensorType.String()
	}

	id = battery.id + "_" + strings.ToLower(strcase.ToSnake(sensorType.String()))

	switch sensorType {
	case typePercentage:
		icon = batteryPercentIcon(value.Value())
		deviceClass = types.SensorDeviceClassBattery
		stateClass = types.StateClassMeasurement
		units = "%"
	case typeTemp:
		deviceClass = types.SensorDeviceClassTemperature
		stateClass = types.StateClassMeasurement
		units = "°C"
	case typeEnergyRate:
		icon = batteryChargeIcon(value.Value())
		deviceClass = types.SensorDeviceClassPower
		stateClass = types.StateClassMeasurement
		units = "W"
	case typeEnergy:
		deviceClass = types.SensorDeviceClassEnergyStorage
		stateClass = types.StateClassMeasurement
		units = "Wh"
	case typeVoltage:
		deviceClass = types.SensorDeviceClassVoltage
		stateClass = types.StateClassMeasurement
		units = "V"
	default:
		icon = batteryIcon
	}

	return sensor.NewSensor(
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithDeviceClass(deviceClass),
		sensor.WithStateClass(stateClass),
		sensor.WithUnits(units),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithValue(generateSensorState(sensorType, value.Value())),
			sensor.WithAttributes(generateSensorAttributes(sensorType, battery)),
		),
	)
}

// generateSensorState will take the raw value (from D-Bus) and format it as
// appropriate for the battery sensor type.
func generateSensorState(sensorType sensorType, value any) any {
	if value == nil {
		return sensor.StateUnknown
	}

	switch sensorType {
	case typeVoltage, typeTemp, typeEnergy, typeEnergyRate, typePercentage:
		if value, ok := value.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case typeState:
		if value, ok := value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return chargingState(value).String()
		}
	case typeLevel:
		if value, ok := value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return level(value).String()
		}
	default:
		if value, ok := value.(string); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	}
}

// generateSensorAttributes will add some appropriate attributes to certain
// battery sensor types.
func generateSensorAttributes(sensorType sensorType, battery *upowerBattery) map[string]any {
	attributes := make(map[string]any)

	attributes["data_source"] = linux.DataSrcDbus

	switch sensorType {
	case typeEnergyRate:
		var (
			variant         dbus.Variant
			err             error
			voltage, energy float64
		)

		if variant, err = battery.getProp(typeVoltage); err == nil {
			voltage, _ = dbusx.VariantToValue[float64](variant) //nolint:lll,errcheck // its not important if this attribute value is not correct due to errors
		}

		if variant, err = battery.getProp(typeEnergy); err == nil {
			energy, _ = dbusx.VariantToValue[float64](variant) //nolint:lll,errcheck // its not important if this attribute value is not correct due to errors
		}

		attributes["voltage"] = voltage
		attributes["energy"] = energy
	case typePercentage, typeLevel:
		attributes["battery_type"] = battery.battType.String()
	}

	return attributes
}

// batteryPercentIcon takes the percent value of level and returns an
// appropriate icon to represent it.
//
//nolint:mnd
func batteryPercentIcon(v any) string {
	percentage, ok := v.(float64)
	if !ok {
		return batteryIcon + "-unknown"
	}

	if percentage >= 95 {
		return batteryIcon
	}

	return fmt.Sprintf("%s-%d", batteryIcon, int(math.Round(percentage/10)*10))
}

// batteryChargeIcon takes the value of the battery charge and returns an
// appropriate icon to represent it.
func batteryChargeIcon(v any) string {
	energyRate, ok := v.(float64)
	if !ok {
		return batteryIcon
	}

	if math.Signbit(energyRate) {
		return batteryIcon + "-minus"
	}

	return batteryIcon + "-plus"
}
