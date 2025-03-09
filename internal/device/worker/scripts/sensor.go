// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package scripts

import (
	"context"
	"errors"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

var ErrNewSensorEntity = errors.New("could not create sensor entity")

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

func scriptToEntity(ctx context.Context, script ScriptSensor) (*models.Entity, error) {
	var typeOption sensor.Option

	switch script.SensorStateType {
	case "binary":
		typeOption = sensor.AsTypeBinarySensor()
	default:
		typeOption = sensor.AsTypeSensor()
	}

	scriptSensor, err := sensor.NewSensor(ctx,
		sensor.WithName(script.SensorName),
		sensor.WithID(strcase.ToSnake(script.SensorName)),
		sensor.WithUnits(script.SensorUnits),
		sensor.WithDeviceClass(script.DeviceClass()),
		sensor.WithStateClass(script.StateClass()),
		sensor.WithIcon(script.Icon()),
		sensor.WithAttributes(script.Attributes()),
		sensor.WithState(script.SensorState),
		typeOption,
	)
	if err != nil {
		return nil, errors.Join(ErrNewSensorEntity, err)
	}

	return &scriptSensor, nil
}

func (s *ScriptSensor) Icon() string {
	if s.SensorIcon == "" {
		return "mdi:script"
	}

	return s.SensorIcon
}

func (s *ScriptSensor) DeviceClass() class.SensorDeviceClass {
	for d := class.SensorClassMin + 1; d <= class.BinaryClassMax; d++ {
		if s.SensorDeviceClass == d.String() {
			return d
		}
	}

	return 0
}

func (s *ScriptSensor) StateClass() class.SensorStateClass {
	switch s.SensorStateClass {
	case "measurement":
		return class.StateMeasurement
	case "total":
		return class.StateTotal
	case "total_increasing":
		return class.StateTotalIncreasing
	default:
		return class.StateClassMin
	}
}

func (s *ScriptSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	if s.SensorAttributes != nil {
		attributes["extra_attributes"] = s.SensorAttributes
	}

	return attributes
}
