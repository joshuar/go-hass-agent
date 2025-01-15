// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package sensor

import (
	"fmt"
	"maps"

	"github.com/joshuar/go-hass-agent/internal/components/validation"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	StateUnknown = "Unknown"
)

type Option[T any] func(T) T

type requestMetadata struct {
	RetryRequest bool
}

type State struct {
	Value      any            `json:"state" validate:"required"`
	Attributes map[string]any `json:"attributes,omitempty" validate:"omitempty"`
	Icon       string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
}

// WithValue assigns a value to the sensor.
func WithValue(value any) Option[State] {
	return func(state State) State {
		state.Value = value
		return state
	}
}

// WithAttributes sets the additional attributes for the sensor.
func WithAttributes(attributes map[string]any) Option[State] {
	return func(state State) State {
		if state.Attributes != nil {
			maps.Copy(state.Attributes, attributes)
		} else {
			state.Attributes = attributes
		}

		return state
	}
}

// WithAttribute sets the given additional attribute to the given value.
func WithAttribute(name string, value any) Option[State] {
	return func(state State) State {
		if state.Attributes == nil {
			state.Attributes = make(map[string]any)
		}

		state.Attributes[name] = value

		return state
	}
}

// WithDataSourceAttribute will set the "data_source" additional attribute to
// the given value.
func WithDataSourceAttribute(source string) Option[State] {
	return func(state State) State {
		if state.Attributes == nil {
			state.Attributes = make(map[string]any)
		}

		state.Attributes["data_source"] = source

		return state
	}
}

// WithIcon sets the sensor icon.
func WithIcon(icon string) Option[State] {
	return func(state State) State {
		state.Icon = icon
		return state
	}
}

// UpdateValue will update the sensor state with the given value.
func (s *State) UpdateValue(value any) {
	s.Value = value
}

// UpdateIcon will update the sensor icon with the given value.
func (s *State) UpdateIcon(icon string) {
	s.Icon = icon
}

// UpdateAttribute will set the given attribute to the given value.
func (s *State) UpdateAttribute(key string, value any) {
	if s.Attributes == nil {
		s.Attributes = make(map[string]any)
	}

	s.Attributes[key] = value
}

func (s *State) Validate() error {
	err := validation.Validate.Struct(s)
	if err != nil {
		return fmt.Errorf("sensor state is invalid: %s", validation.ParseValidationErrors(err))
	}

	return nil
}

type Entity struct {
	*State
	requestMetadata
	ID          string            `json:"unique_id" validate:"required"`
	Name        string            `json:"name" validate:"required"`
	Units       string            `json:"unit_of_measurement,omitempty" validate:"omitempty"`
	EntityType  types.SensorType  `json:"type" validate:"omitempty"`
	DeviceClass types.DeviceClass `json:"device_class,omitempty" validate:"omitempty"`
	StateClass  types.StateClass  `json:"state_class,omitempty" validate:"omitempty"`
	Category    types.Category    `json:"entity_category,omitempty" validate:"omitempty"`
}

// WithState sets the sensor state options. This is useful on entity
// creation to set an initial state.
func WithState(options ...Option[State]) Option[Entity] {
	return func(entity Entity) Entity {
		state := State{}

		for _, option := range options {
			state = option(state)
		}

		entity.State = &state

		return entity
	}
}

// WithName sets the friendly name for the sensor entity.
func WithName(name string) Option[Entity] {
	return func(entity Entity) Entity {
		entity.Name = name
		return entity
	}
}

// WithID sets the entity ID of the sensor.
func WithID(id string) Option[Entity] {
	return func(entity Entity) Entity {
		entity.ID = id
		return entity
	}
}

// AsTypeSensor ensures the sensor is treated as a Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/
func AsTypeSensor() Option[Entity] {
	return func(entity Entity) Entity {
		entity.EntityType = types.Sensor
		return entity
	}
}

// AsTypeBinarySensor ensures the sensor is treated as a Binary Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/binary-sensor
func AsTypeBinarySensor() Option[Entity] {
	return func(entity Entity) Entity {
		entity.EntityType = types.BinarySensor
		return entity
	}
}

// WithUnits defines the native unit of measurement of the sensor entity.
func WithUnits(units string) Option[Entity] {
	return func(entity Entity) Entity {
		entity.Units = units
		return entity
	}
}

// WithDeviceClass sets the device class of the sensor entity.
//
// For type Sensor: https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes
//
// For type Binary Sensor: https://developers.home-assistant.io/docs/core/entity/binary-sensor#available-device-classes
func WithDeviceClass(class types.DeviceClass) Option[Entity] {
	return func(entity Entity) Entity {
		entity.DeviceClass = class
		return entity
	}
}

// WithStateClass sets the state class of the sensor entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes
func WithStateClass(class types.StateClass) Option[Entity] {
	return func(entity Entity) Entity {
		entity.StateClass = class
		return entity
	}
}

// AsDiagnostic sets the sensor entity as a diagnostic. This will ensure it will
// be grouped under a diagnostic header in the Home Assistant UI.
func AsDiagnostic() Option[Entity] {
	return func(entity Entity) Entity {
		entity.Category = types.CategoryDiagnostic
		return entity
	}
}

// WithRequestRetry flags that any API requests for this entity should be
// retried.
func WithRequestRetry(value bool) Option[Entity] {
	return func(e Entity) Entity {
		e.RetryRequest = value
		return e
	}
}

// NewSensor provides a way to build a sensor entity with the given options.
func NewSensor(options ...Option[Entity]) Entity {
	sensor := Entity{}

	for _, option := range options {
		sensor = option(sensor)
	}

	return sensor
}

// UpdateState will set the state of the entity as per the given options. This can
// be used on an existing Entity to "update" the state. Note that any existing
// state will be reset and only the new options will be applied.
func (e *Entity) UpdateState(options ...Option[State]) {
	state := State{}
	for _, option := range options {
		state = option(state)
	}

	e.State = &state
}

func (e *Entity) Validate() error {
	err := validation.Validate.Struct(e)
	if err != nil {
		return fmt.Errorf("sensor is invalid: %s", validation.ParseValidationErrors(err))
	}

	return nil
}

// Location represents the location information that can be sent to HA to
// update the location of the agent. This is exposed so that device code can
// create location requests directly, as Home Assistant handles these
// differently from other sensors.
type Location struct {
	Gps              []float64 `json:"gps" validate:"required"`
	GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
	Battery          int       `json:"battery,omitempty"`
	Speed            int       `json:"speed,omitempty"`
	Altitude         int       `json:"altitude,omitempty"`
	Course           int       `json:"course,omitempty"`
	VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
}

func (l *Location) Validate() error {
	err := validation.Validate.Struct(l)
	if err != nil {
		return fmt.Errorf("location is invalid: %s", validation.ParseValidationErrors(err))
	}

	return nil
}
