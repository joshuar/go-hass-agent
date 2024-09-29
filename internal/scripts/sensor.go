// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package scripts

import (
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

//nolint:tagalign
type ScriptSensor struct {
	SensorState       any    `json:"sensor_state" yaml:"sensor_state" toml:"sensor_state"`
	SensorAttributes  any    `json:"sensor_attributes,omitempty" yaml:"sensor_attributes,omitempty" toml:"sensor_attributes,omitempty"`
	SensorName        string `json:"sensor_name" yaml:"sensor_name" toml:"sensor_name"`
	SensorIcon        string `json:"sensor_icon,omitempty" yaml:"sensor_icon,omitempty" toml:"sensor_icon,omitempty"`
	SensorDeviceClass string `json:"sensor_device_class,omitempty" yaml:"sensor_device_class,omitempty" toml:"sensor_device_class,omitempty"`
	SensorStateClass  string `json:"sensor_state_class,omitempty" yaml:"sensor_state_class,omitempty" toml:"sensor_state_class,omitempty"`
	SensorStateType   string `json:"sensor_type,omitempty" yaml:"sensor_type,omitempty" toml:"sensor_type,omitempty"`
	SensorUnits       string `json:"sensor_units,omitempty" yaml:"sensor_units,omitempty" toml:"sensor_units,omitempty"`
}

func scriptToEntity(script ScriptSensor) sensor.Entity {
	return sensor.Entity{
		Name:        script.SensorName,
		Units:       script.SensorUnits,
		DeviceClass: script.DeviceClass(),
		StateClass:  script.StateClass(),
		EntityState: &sensor.EntityState{
			State:      script.SensorState,
			ID:         strcase.ToSnake(script.SensorName),
			Icon:       script.Icon(),
			Attributes: script.Attributes(),
			EntityType: script.SensorType(),
		},
	}
}

func (s *ScriptSensor) Icon() string {
	if s.SensorIcon == "" {
		return "mdi:script"
	}

	return s.SensorIcon
}

func (s *ScriptSensor) SensorType() types.SensorClass {
	switch s.SensorStateType {
	case "binary":
		return types.BinarySensor
	default:
		return types.Sensor
	}
}

func (s *ScriptSensor) DeviceClass() types.DeviceClass {
	for d := types.SensorDeviceClassApparentPower; d <= types.SensorDeviceClassWindSpeed; d++ {
		if s.SensorDeviceClass == d.String() {
			return d
		}
	}

	return 0
}

func (s *ScriptSensor) StateClass() types.StateClass {
	switch s.SensorStateClass {
	case "measurement":
		return types.StateClassMeasurement
	case "total":
		return types.StateClassTotal
	case "total_increasing":
		return types.StateClassTotalIncreasing
	default:
		return 0
	}
}

func (s *ScriptSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	if s.SensorAttributes != nil {
		attributes["extra_attributes"] = s.SensorAttributes
	}

	return attributes
}
