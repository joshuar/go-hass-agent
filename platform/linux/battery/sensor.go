// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package battery

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const unknownValue = "Unknown"

// newBatterySensor creates a new sensor for Home Assistant from a battery
// property.
func newBatterySensor(
	ctx context.Context,
	battery *upowerBattery,
	sensorType sensorType,
	value dbus.Variant,
) models.Entity {
	var (
		name, id, icon, units string
		deviceClass           models.SensorDeviceClass
		stateClass            models.SensorStateClass
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
		deviceClass = models.SensorClassBattery
		stateClass = models.StateMeasurement
		units = "%"
	case typeTemp:
		deviceClass = models.SensorClassTemperature
		stateClass = models.StateMeasurement
		units = "°C"
	case typeEnergyRate:
		icon = batteryChargeIcon(value.Value())
		deviceClass = models.SensorClassPower
		stateClass = models.StateMeasurement
		units = "W"
	case typeEnergy:
		deviceClass = models.SensorClassEnergyStorage
		stateClass = models.StateMeasurement
		units = "Wh"
	case typeVoltage:
		deviceClass = models.SensorClassVoltage
		stateClass = models.StateMeasurement
		units = "V"
	default:
		icon = batteryIcon
	}

	return models.NewSensor(ctx,
		models.WithName(name),
		models.WithID(id),
		models.WithDeviceClass(deviceClass),
		models.WithStateClass(stateClass),
		models.WithUnits(units),
		models.AsDiagnostic(),
		models.WithIcon(icon),
		models.WithState(generateSensorState(sensorType, value.Value())),
		models.WithAttributes(generateSensorAttributes(sensorType, battery)),
	)
}

// generateSensorState will take the raw value (from D-Bus) and format it as
// appropriate for the battery sensor type.
func generateSensorState(sensorType sensorType, value any) any {
	if value == nil {
		return unknownValue
	}

	switch sensorType {
	case typeVoltage, typeTemp, typeEnergy, typeEnergyRate, typePercentage:
		value, ok := value.(float64)
		if !ok {
			return unknownValue
		}

		return value
	case typeState:
		value, ok := value.(uint32)
		if !ok {
			return unknownValue
		}

		return chargingState(value).String()
	case typeLevel:
		value, ok := value.(uint32)
		if !ok {
			return unknownValue
		}

		return level(value).String()
	default:
		value, ok := value.(string)
		if !ok {
			return unknownValue
		}

		return value
	}
}

// generateSensorAttributes will add some appropriate attributes to certain
// battery sensor types.
func generateSensorAttributes(sensorType sensorType, battery *upowerBattery) map[string]any {
	attributes := make(map[string]any)

	attributes["data_source"] = linux.DataSrcDBus

	switch sensorType {
	case typeEnergyRate:
		var (
			variant         dbus.Variant
			err             error
			voltage, energy float64
		)

		if variant, err = battery.getProp(typeVoltage); err == nil {
			voltage, _ = dbusx.VariantToValue[float64](
				variant,
			) //nolint:lll // its not important if this attribute value is not correct due to errors
		}

		if variant, err = battery.getProp(typeEnergy); err == nil {
			energy, _ = dbusx.VariantToValue[float64](
				variant,
			) //nolint:lll // its not important if this attribute value is not correct due to errors
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
