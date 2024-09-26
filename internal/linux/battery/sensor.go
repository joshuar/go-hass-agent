// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
func newBatterySensor(battery *upowerBattery, sensorType batterySensor, value dbus.Variant) sensor.Entity {
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
	case battPercentage:
		icon = batteryPercentIcon(value.Value())
		deviceClass = types.DeviceClassBattery
		stateClass = types.StateClassMeasurement
		units = "%"
	case battTemp:
		deviceClass = types.DeviceClassTemperature
		stateClass = types.StateClassMeasurement
		units = "°C"
	case battEnergyRate:
		icon = batteryChargeIcon(value.Value())
		deviceClass = types.DeviceClassPower
		stateClass = types.StateClassMeasurement
		units = "W"
	default:
		icon = batteryIcon
	}

	return sensor.Entity{
		Name:        name,
		Category:    types.CategoryDiagnostic,
		DeviceClass: deviceClass,
		StateClass:  stateClass,
		Units:       units,
		EntityState: &sensor.EntityState{
			ID:         id,
			Icon:       icon,
			State:      generateSensorState(sensorType, value.Value()),
			Attributes: generateSensorAttributes(sensorType, battery),
		},
	}
}

func generateSensorState(sensorType batterySensor, value any) any {
	if value == nil {
		return sensor.StateUnknown
	}

	switch sensorType {
	case battVoltage, battTemp, battEnergy, battEnergyRate, battPercentage:
		if value, ok := value.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case battState:
		if value, ok := value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return battChargeState(value).String()
		}
	case battLevel:
		if value, ok := value.(uint32); !ok {
			return sensor.StateUnknown
		} else {
			return batteryLevel(value).String()
		}
	default:
		if value, ok := value.(string); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	}
}

//nolint:exhaustive,errcheck
func generateSensorAttributes(sensorType batterySensor, battery *upowerBattery) map[string]any {
	attributes := make(map[string]any)

	attributes["data_source"] = linux.DataSrcDbus

	switch sensorType {
	case battEnergyRate:
		var (
			variant         dbus.Variant
			err             error
			voltage, energy float64
		)

		if variant, err = battery.getProp(battVoltage); err == nil {
			voltage, _ = dbusx.VariantToValue[float64](variant)
		}

		if variant, err = battery.getProp(battEnergy); err == nil {
			energy, _ = dbusx.VariantToValue[float64](variant)
		}

		attributes["voltage"] = voltage
		attributes["energy"] = energy
	case battPercentage, battLevel:
		attributes["battery_type"] = battery.battType.String()
	}

	return attributes
}

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
