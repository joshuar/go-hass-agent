// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package sensor provides a method and options for creating model.Sensor
// objects wrapped as a model.Entity.
package sensor

import (
	"context"
	"errors"
	"log/slog"
	"maps"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
)

var ErrNewSensor = errors.New("could not create new sensor")

type Option models.Option[*models.Sensor]

// WithState assigns a state value to the Sensor.
func WithState(value any) Option {
	return func(s *models.Sensor) error {
		s.State = value
		return nil
	}
}

// WithAttributes option sets the additional attributes for the sensor.
func WithAttributes(attributes map[string]any) Option {
	return func(s *models.Sensor) error {
		if attributes == nil {
			return nil
		}

		maps.Copy(s.Attributes, attributes)

		return nil
	}
}

// WithAttribute sets the given additional attribute to the given value.
func WithAttribute(name string, value any) Option {
	return func(s *models.Sensor) error {
		s.Attributes[name] = value
		return nil
	}
}

// WithDataSourceAttribute will set the "data_source" additional attribute to
// the given value.
func WithDataSourceAttribute(source string) Option {
	return func(s *models.Sensor) error {
		return WithAttribute("data_source", source)(s)
	}
}

// WithIcon sets the sensor icon.
func WithIcon(icon string) Option {
	return func(s *models.Sensor) error {
		if icon != "" {
			s.Icon = &icon
		}

		return nil
	}
}

// WithName sets the friendly name for the sensor entity.
func WithName(name string) Option {
	return func(s *models.Sensor) error {
		s.Name = name
		return nil
	}
}

// WithID sets the entity ID of the sensor.
func WithID(id string) Option {
	return func(s *models.Sensor) error {
		s.UniqueID = id
		return nil
	}
}

// AsTypeSensor ensures the sensor is treated as a Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/
func AsTypeSensor() Option {
	return func(s *models.Sensor) error {
		s.Type = models.SensorTypeSensor
		return nil
	}
}

// AsTypeBinarySensor ensures the sensor is treated as a Binary Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/binary-sensor
func AsTypeBinarySensor() Option {
	return func(s *models.Sensor) error {
		s.Type = models.SensorTypeBinarySensor
		return nil
	}
}

// WithUnits defines the native unit of measurement of the sensor entity.
func WithUnits(units string) Option {
	return func(s *models.Sensor) error {
		if units != "" {
			s.UnitOfMeasurement = &units
		}

		return nil
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
	return func(s *models.Sensor) error {
		if deviceClass.Valid() {
			str := deviceClass.String()
			s.DeviceClass = &str
		}

		return nil
	}
}

// WithStateClass option sets the state class of the sensor entity. If the given
// state class is an invalid value, it is ignored.
//
// https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes
func WithStateClass(stateClass class.SensorStateClass) Option {
	return func(s *models.Sensor) error {
		if stateClass.Valid() {
			str := stateClass.String()
			s.StateClass = &str
		}

		return nil
	}
}

// WithCategory option sets the entity category explicitly to the value given.
// If the value is invalid or empty, it is ignored.
func WithCategory(category models.EntityCategory) Option {
	return func(s *models.Sensor) error {
		if category != "" {
			s.EntityCategory = &category
		}

		return nil
	}
}

// AsDiagnostic sets the sensor entity as a diagnostic. This will ensure it will
// be grouped under a diagnostic header in the Home Assistant UI.
func AsDiagnostic() Option {
	return func(s *models.Sensor) error {
		category := models.Diagnostic
		s.EntityCategory = &category

		return nil
	}
}

// AsRetryableRequest sets a flag on the sensor that indicates the requests sent
// to Home Assistant related to this sensor should be retried.
func AsRetryableRequest(value bool) Option {
	return func(s *models.Sensor) error {
		s.Retryable = value
		return nil
	}
}

// NewSensor provides a way to build a sensor entity with the given options.
func NewSensor(ctx context.Context, options ...Option) (models.Entity, error) {
	sensor := models.Sensor{
		Attributes: make(models.Attributes),
	}

	for _, option := range options {
		if err := option(&sensor); err != nil {
			logging.FromContext(ctx).Warn("Could not set sensor option.", slog.Any("error", err))
		}
	}

	if sensor.Type == "" {
		if err := AsTypeSensor()(&sensor); err != nil {
			logging.FromContext(ctx).Warn("Could not set sensor option.", slog.Any("error", err))
		}
	}

	entity := models.Entity{}

	err := entity.FromSensor(sensor)
	if err != nil {
		return entity, errors.Join(ErrNewSensor, err)
	}

	return entity, nil
}
