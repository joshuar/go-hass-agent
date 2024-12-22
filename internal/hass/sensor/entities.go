// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package sensor

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/validation"
)

const (
	StateUnknown = "Unknown"

	requestTypeRegisterSensor = "register_sensor"
	requestTypeUpdateSensor   = "update_sensor_states"
	requestTypeLocation       = "update_location"
)

type Option[T any] func(T) T

type Request struct {
	Data        any    `json:"data"`
	RequestType string `json:"type"`
}

type RequestMetadata struct {
	RetryRequest bool
}

type State struct {
	Value      any              `json:"state" validate:"required"`
	Attributes map[string]any   `json:"attributes,omitempty" validate:"omitempty"`
	Icon       string           `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
	ID         string           `json:"unique_id" validate:"required"`
	EntityType types.SensorType `json:"type" validate:"omitempty"`
	RequestMetadata
}

// WithValue assigns a value to the sensor.
func WithValue(value any) Option[State] {
	return func(s State) State {
		s.Value = value
		return s
	}
}

// WithAttributes sets the additional attributes for the sensor.
func WithAttributes(attributes map[string]any) Option[State] {
	return func(s State) State {
		if s.Attributes != nil {
			maps.Copy(s.Attributes, attributes)
		} else {
			s.Attributes = attributes
		}
		return s
	}
}

// WithAttribute sets the given additional attribute to the given value.
func WithAttribute(name string, value any) Option[State] {
	return func(s State) State {
		if s.Attributes == nil {
			s.Attributes = make(map[string]any)
		}

		s.Attributes[name] = value

		return s
	}
}

// WithDataSourceAttribute will set the "data_source" additional attribute to
// the given value.
func WithDataSourceAttribute(source string) Option[State] {
	return func(s State) State {
		if s.Attributes == nil {
			s.Attributes = make(map[string]any)
		}

		s.Attributes["data_source"] = source

		return s
	}
}

// WithIcon sets the sensor icon.
func WithIcon(icon string) Option[State] {
	return func(s State) State {
		s.Icon = icon
		return s
	}
}

// WithID sets the entity ID of the sensor.
func WithID(id string) Option[State] {
	return func(s State) State {
		s.ID = id
		return s
	}
}

// AsTypeSensor ensures the sensor is treated as a Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/
func AsTypeSensor() Option[State] {
	return func(s State) State {
		s.EntityType = types.Sensor
		return s
	}
}

// AsTypeBinarySensor ensures the sensor is treated as a Binary Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/binary-sensor
func AsTypeBinarySensor() Option[State] {
	return func(s State) State {
		s.EntityType = types.BinarySensor
		return s
	}
}

// UpdateValue will update the sensor state with the given value.
func (e *State) UpdateValue(value any) {
	e.Value = value
}

// UpdateIcon will update the sensor icon with the given value.
func (e *State) UpdateIcon(icon string) {
	e.Icon = icon
}

// UpdateAttribute will set the given attribute to the given value.
func (e *State) UpdateAttribute(key string, value any) {
	if e.Attributes == nil {
		e.Attributes = make(map[string]any)
	}
	e.Attributes[key] = value
}

func (s *State) Validate() error {
	err := validation.Validate.Struct(s)
	if err != nil {
		return fmt.Errorf("sensor state is invalid: %s", validation.ParseValidationErrors(err))
	}

	return nil
}

func (s *State) RequestBody() any {
	return &Request{
		RequestType: requestTypeUpdateSensor,
		Data:        s,
	}
}

func (s *State) Retry() bool {
	return s.RetryRequest
}

//nolint:wrapcheck
func (s *State) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		State      any            `json:"state" validate:"required"`
		Attributes map[string]any `json:"attributes,omitempty" validate:"omitempty"`
		Icon       string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
		ID         string         `json:"unique_id" validate:"required"`
		EntityType string         `json:"type" validate:"omitempty"`
	}{
		State:      s.Value,
		Attributes: s.Attributes,
		Icon:       s.Icon,
		ID:         s.ID,
		EntityType: s.EntityType.String(),
	})
}

type Entity struct {
	*State
	Name        string            `json:"name" validate:"required"`
	Units       string            `json:"unit_of_measurement,omitempty" validate:"omitempty"`
	DeviceClass types.DeviceClass `json:"device_class,omitempty" validate:"omitempty"`
	StateClass  types.StateClass  `json:"state_class,omitempty" validate:"omitempty"`
	Category    types.Category    `json:"entity_category,omitempty" validate:"omitempty"`
}

// WithState sets the sensor state options. This is useful on entity
// creation to set an intial state.
func WithState(options ...Option[State]) Option[Entity] {
	return func(e Entity) Entity {
		state := State{}

		for _, option := range options {
			state = option(state)
		}

		e.State = &state

		return e
	}
}

// WithName sets the friendly name for the sensor entity.
func WithName(name string) Option[Entity] {
	return func(e Entity) Entity {
		e.Name = name
		return e
	}
}

// WithUnits defines the native unit of measurement of the sensor entity.
func WithUnits(units string) Option[Entity] {
	return func(e Entity) Entity {
		e.Units = units
		return e
	}
}

// WithDeviceClass sets the device class of the sensor entity.
//
// For type Sensor: https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes
//
// For type Binary Sensor: https://developers.home-assistant.io/docs/core/entity/binary-sensor#available-device-classes
func WithDeviceClass(class types.DeviceClass) Option[Entity] {
	return func(e Entity) Entity {
		e.DeviceClass = class
		return e
	}
}

// WithStateClass sets the state class of the sensor entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes
func WithStateClass(class types.StateClass) Option[Entity] {
	return func(e Entity) Entity {
		e.StateClass = class
		return e
	}
}

// AsDiagnostic sets the sensor entity as a diagnostic. This will ensure it will
// be grouped under a diagnostic header in the Home Assistant UI.
func AsDiagnostic() Option[Entity] {
	return func(e Entity) Entity {
		e.Category = types.CategoryDiagnostic
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

func (e *Entity) RequestBody() any {
	return &Request{
		RequestType: requestTypeRegisterSensor,
		Data:        e,
	}
}

func (e *Entity) Retry() bool {
	return e.RetryRequest
}

//nolint:wrapcheck
func (e *Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		State       any            `json:"state" validate:"required"`
		Attributes  map[string]any `json:"attributes,omitempty" validate:"omitempty"`
		Icon        string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
		ID          string         `json:"unique_id" validate:"required"`
		EntityType  string         `json:"type" validate:"omitempty"`
		Name        string         `json:"name" validate:"required"`
		Units       string         `json:"unit_of_measurement,omitempty" validate:"omitempty"`
		DeviceClass string         `json:"device_class,omitempty" validate:"omitempty"`
		StateClass  string         `json:"state_class,omitempty" validate:"omitempty"`
		Category    string         `json:"entity_category,omitempty" validate:"omitempty"`
	}{
		State:       e.Value,
		Attributes:  e.Attributes,
		Icon:        e.Icon,
		ID:          e.ID,
		EntityType:  e.EntityType.String(),
		Name:        e.Name,
		Units:       e.Units,
		DeviceClass: e.DeviceClass.String(),
		StateClass:  e.StateClass.String(),
		Category:    e.Category.String(),
	})
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

func (l *Location) RequestBody() any {
	return &Request{
		RequestType: requestTypeLocation,
		Data:        l,
	}
}

func (l *Location) Retry() bool {
	return false
}
