// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package sensor provides a method and options for creating model.Sensor
// objects wrapped as a model.Entity.
package sensor

import (
	"context"
	"maps"

	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
)

// Option is a functional option for a sensor.
type Option models.Option[*models.Sensor]

// WithState assigns a state value to the Sensor.
func WithState(value any) Option {
	return func(s *models.Sensor) {
		s.State = value
	}
}

// WithAttributes option sets the additional attributes for the sensor.
func WithAttributes(attributes map[string]any) Option {
	return func(s *models.Sensor) {
		if attributes == nil {
			return
		}
		maps.Copy(s.Attributes, attributes)
	}
}

// WithAttribute sets the given additional attribute to the given value.
func WithAttribute(name string, value any) Option {
	return func(s *models.Sensor) {
		s.Attributes[name] = value
	}
}

// WithDataSourceAttribute will set the "data_source" additional attribute to
// the given value.
func WithDataSourceAttribute(source string) Option {
	return func(s *models.Sensor) {
		WithAttribute("data_source", source)(s)
	}
}

// WithIcon sets the sensor icon.
func WithIcon(icon string) Option {
	return func(s *models.Sensor) {
		if icon != "" {
			s.Icon = &icon
		}
	}
}

// WithName sets the friendly name for the sensor entity.
func WithName(name string) Option {
	return func(s *models.Sensor) {
		s.Name = name
	}
}

// WithID sets the entity ID of the sensor.
func WithID(id string) Option {
	return func(s *models.Sensor) {
		s.UniqueID = id
	}
}

// AsTypeSensor ensures the sensor is treated as a Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/
func AsTypeSensor() Option {
	return func(s *models.Sensor) {
		s.Type = models.SensorTypeSensor
	}
}

// AsTypeBinarySensor ensures the sensor is treated as a Binary Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/binary-sensor
func AsTypeBinarySensor() Option {
	return func(s *models.Sensor) {
		s.Type = models.SensorTypeBinarySensor
	}
}

// WithUnits defines the native unit of measurement of the sensor entity.
func WithUnits(units string) Option {
	return func(s *models.Sensor) {
		if units != "" {
			s.UnitOfMeasurement = &units
		}
	}
}

// WithDeviceClass sets the device class of the sensor entity.
//
// For type Sensor:
//
// https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes
//
// For type Binary Sensor:
//
// https://developers.home-assistant.io/docs/core/entity/binary-sensor#available-device-classes
func WithDeviceClass(deviceClass class.SensorDeviceClass) Option {
	return func(s *models.Sensor) {
		if deviceClass.Valid() {
			str := deviceClass.String()
			s.DeviceClass = &str
		}
	}
}

// WithStateClass option sets the state class of the sensor entity. If the given
// state class is an invalid value, it is ignored.
//
// https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes
func WithStateClass(stateClass class.SensorStateClass) Option {
	return func(s *models.Sensor) {
		if stateClass.Valid() {
			str := stateClass.String()
			s.StateClass = &str
		}
	}
}

// WithCategory option sets the entity category explicitly to the value given.
// If the value is invalid or empty, it is ignored.
func WithCategory(category models.EntityCategory) Option {
	return func(s *models.Sensor) {
		if category != "" {
			s.EntityCategory = &category
		}
	}
}

// AsDiagnostic sets the sensor entity as a diagnostic. This will ensure it will
// be grouped under a diagnostic header in the Home Assistant UI.
func AsDiagnostic() Option {
	return func(s *models.Sensor) {
		category := models.Diagnostic
		s.EntityCategory = &category
	}
}

// AsRetryableRequest sets a flag on the sensor that indicates the requests sent
// to Home Assistant related to this sensor should be retried.
func AsRetryableRequest(value bool) Option {
	return func(s *models.Sensor) {
		s.Retryable = value
	}
}

// NewSensor provides a way to build a sensor entity with the given options.
func NewSensor(ctx context.Context, options ...Option) models.Entity {
	sensor := models.Sensor{
		Attributes: make(models.Attributes),
	}

	for _, option := range options {
		option(&sensor)
	}

	if sensor.Type == "" {
		AsTypeSensor()(&sensor)
	}

	entity := models.Entity{}
	entity.FromSensor(sensor)
	return entity
}
