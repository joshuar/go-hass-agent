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

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const unknownValue = "Unknown"

// newBatterySensor creates a new sensor for Home Assistant from a battery
// property.
func newBatterySensor(ctx context.Context, battery *upowerBattery, sensorType sensorType, value dbus.Variant) (models.Entity, error) {
	var (
		name, id, icon, units string
		deviceClass           class.SensorDeviceClass
		stateClass            class.SensorStateClass
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
		deviceClass = class.SensorClassBattery
		stateClass = class.StateMeasurement
		units = "%"
	case typeTemp:
		deviceClass = class.SensorClassTemperature
		stateClass = class.StateMeasurement
		units = "°C"
	case typeEnergyRate:
		icon = batteryChargeIcon(value.Value())
		deviceClass = class.SensorClassPower
		stateClass = class.StateMeasurement
		units = "W"
	case typeEnergy:
		deviceClass = class.SensorClassEnergyStorage
		stateClass = class.StateMeasurement
		units = "Wh"
	case typeVoltage:
		deviceClass = class.SensorClassVoltage
		stateClass = class.StateMeasurement
		units = "V"
	default:
		icon = batteryIcon
	}

	entity, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithDeviceClass(deviceClass),
		sensor.WithStateClass(stateClass),
		sensor.WithUnits(units),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(generateSensorState(sensorType, value.Value())),
		sensor.WithAttributes(generateSensorAttributes(sensorType, battery)),
	)
	if err != nil {
		return entity, fmt.Errorf("could not generate %s battery sensor: %w", name, err)
	}

	return entity, nil
}

// generateSensorState will take the raw value (from D-Bus) and format it as
// appropriate for the battery sensor type.
func generateSensorState(sensorType sensorType, value any) any {
	if value == nil {
		return unknownValue
	}

	switch sensorType {
	case typeVoltage, typeTemp, typeEnergy, typeEnergyRate, typePercentage:
		if value, ok := value.(float64); !ok {
			return unknownValue
		} else {
			return value
		}
	case typeState:
		if value, ok := value.(uint32); !ok {
			return unknownValue
		} else {
			return chargingState(value).String()
		}
	case typeLevel:
		if value, ok := value.(uint32); !ok {
			return unknownValue
		} else {
			return level(value).String()
		}
	default:
		if value, ok := value.(string); !ok {
			return unknownValue
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
