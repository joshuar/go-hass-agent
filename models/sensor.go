// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strconv"

	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/validation"
)

var (
	ErrMarshalSensor   = errors.New("could not marshal entity data")
	ErrUnmarshalSensor = errors.New("could not unmarshal entity data")
	ErrInvalidSensor   = errors.New("sensor data is invalid")
)

// Valid returns a boolean indicating whether the SensorState date is valid.
func (s *SensorState) Valid() (bool, error) {
	if err := validation.ValidateStruct(s); err != nil {
		return false, fmt.Errorf("%w: %s", ErrInvalidSensor, err)
	}

	return true, nil
}

// Valid returns a boolean indicating whether the SensorRegistration data is valid.
func (s *SensorRegistration) Valid() (bool, error) {
	if err := validation.ValidateStruct(s); err != nil {
		return false, fmt.Errorf("%w: %s", ErrInvalidSensor, err)
	}

	return true, nil
}

// SensorOption is a functional option for a sensor.
type SensorOption Option[*Sensor]

// WithState option assigns a state value to the Sensor.
func WithState(value any) SensorOption {
	return func(s *Sensor) {
		s.State = value
	}
}

// WithAttributes option sets the additional attributes for the sensor.
func WithAttributes(attributes map[string]any) SensorOption {
	return func(s *Sensor) {
		if attributes == nil {
			return
		}
		maps.Copy(s.Attributes, attributes)
	}
}

// WithAttribute option sets the given additional attribute to the given value.
func WithAttribute(name string, value any) SensorOption {
	return func(s *Sensor) {
		s.Attributes[name] = value
	}
}

// WithDataSourceAttribute option will set the "data_source" additional attribute to
// the given value.
func WithDataSourceAttribute(source string) SensorOption {
	return func(s *Sensor) {
		WithAttribute("data_source", source)(s)
	}
}

// WithIcon option sets the sensor icon.
func WithIcon(icon string) SensorOption {
	return func(s *Sensor) {
		if icon != "" {
			s.Icon = icon
		}
	}
}

// WithName option sets the friendly name for the sensor entity.
func WithName(name string) SensorOption {
	return func(s *Sensor) {
		s.Name = name
	}
}

// WithID option sets the entity ID of the sensor.
func WithID(id string) SensorOption {
	return func(s *Sensor) {
		s.UniqueID = id
	}
}

// AsTypeSensor option ensures the sensor is treated as a Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/sensor/
func AsTypeSensor() SensorOption {
	return func(s *Sensor) {
		s.Type = SensorTypeSensor
	}
}

// AsTypeBinarySensor option ensures the sensor is treated as a Binary Sensor Entity.
// https://developers.home-assistant.io/docs/core/entity/binary-sensor
func AsTypeBinarySensor() SensorOption {
	return func(s *Sensor) {
		s.Type = SensorTypeBinarySensor
	}
}

// WithUnits option defines the native unit of measurement of the sensor entity.
func WithUnits(units string) SensorOption {
	return func(s *Sensor) {
		if units != "" {
			s.UnitOfMeasurement = units
		}
	}
}

// WithDeviceClass option sets the device class of the sensor entity.
//
// For type Sensor:
//
// https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes
//
// For type Binary Sensor:
//
// https://developers.home-assistant.io/docs/core/entity/binary-sensor#available-device-classes
func WithDeviceClass(deviceClass class.SensorDeviceClass) SensorOption {
	return func(s *Sensor) {
		if deviceClass.Valid() {
			str := deviceClass.String()
			s.DeviceClass = str
		}
	}
}

// WithStateClass option sets the state class of the sensor entity. If the given
// state class is an invalid value, it is ignored.
//
// https://developers.home-assistant.io/docs/core/entity/sensor/#available-state-classes
func WithStateClass(stateClass class.SensorStateClass) SensorOption {
	return func(s *Sensor) {
		if stateClass.Valid() {
			str := stateClass.String()
			s.StateClass = str
		}
	}
}

// WithCategory option sets the entity category explicitly to the value given.
// If the value is invalid or empty, it is ignored.
func WithCategory(category EntityCategory) SensorOption {
	return func(s *Sensor) {
		if category != "" {
			s.EntityCategory = category
		}
	}
}

// AsDiagnostic option sets the sensor entity as a diagnostic. This will ensure it will
// be grouped under a diagnostic header in the Home Assistant UI.
func AsDiagnostic() SensorOption {
	return func(s *Sensor) {
		category := EntityCategoryDiagnostic
		s.EntityCategory = category
	}
}

// AsRetryableRequest option sets a flag on the sensor that indicates the requests sent
// to Home Assistant related to this sensor should be retried.
func AsRetryableRequest(value bool) SensorOption {
	return func(s *Sensor) {
		s.Retryable = value
	}
}

// NewSensor provides a way to build a sensor entity with the given options.
func NewSensor(_ context.Context, options ...SensorOption) Entity {
	sensor := Sensor{
		Attributes: make(Attributes),
	}

	for _, option := range options {
		option(&sensor)
	}

	if sensor.Type == "" {
		AsTypeSensor()(&sensor)
	}

	entity := Entity{}
	entity.FromSensor(sensor)
	return entity
}

// // String returns a string representation of a sensor.
// func (s *Sensor) String() string {
// 	var b strings.Builder

// 	fmt.Fprintf(&b, "Name: %s ", s.Name)
// 	fmt.Fprintf(&b, "ID: %s ", s.UniqueID)
// 	fmt.Fprintf(&b, "Name: %s ", s.Name)

// 	if s.UnitOfMeasurement != nil {
// 		fmt.Fprintf(&b, "State: %v %s", s.State, *s.UnitOfMeasurement)
// 	} else {
// 		fmt.Fprintf(&b, "State: %v", s.State)
// 	}

// 	return b.String()
// }

// LogAttributes returns an slog.Group of log attributes for a sensor entity.
func (s *Sensor) LogAttributes() slog.Attr {
	var state string
	if s.UnitOfMeasurement != "" {
		state = fmt.Sprintf("%v %s", s.State, s.UnitOfMeasurement)
	} else {
		state = fmt.Sprintf("%v", s.State)
	}

	return slog.Group("sensor",
		slog.String("name", s.Name),
		slog.String("id", s.UniqueID),
		slog.String("state", state),
	)
}

// AsState returns the Sensor data as a SensorState object, which can be sent to
// Home Assistant as a sensor update request.
func (s *Sensor) AsState() (*SensorState, error) {
	// Marshal the sensor data to json.
	data, err := json.Marshal(s)
	if err != nil {
		return nil, errors.Join(ErrMarshalSensor, err)
	}

	state := SensorState{}

	// Unmarshal the sensor data back into a sensor state.
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, errors.Join(ErrUnmarshalSensor, err)
	}

	if valid, err := state.Valid(); !valid {
		return nil, err
	}

	return &state, nil
}

// AsRegistration returns the Sensor data as a SensorRegistration object, which can be sent to
// Home Assistant as a sensor registration request.
func (s *Sensor) AsRegistration() (*SensorRegistration, error) {
	// Marshal the sensor data to json.
	data, err := json.Marshal(s)
	if err != nil {
		return nil, errors.Join(ErrMarshalSensor, err)
	}

	registration := SensorRegistration{}

	// Unmarshal the sensor data back into a sensor state.
	err = json.Unmarshal(data, &registration)
	if err != nil {
		return nil, errors.Join(ErrUnmarshalSensor, err)
	}

	if valid, err := registration.Valid(); !valid {
		return nil, err
	}

	return &registration, nil
}

func (s *Sensor) FormatState() string {
	var stateValue string
	switch value := s.State.(type) {
	case string:
		stateValue = value
	case int:
		stateValue = strconv.Itoa(value)
	case int64:
		stateValue = strconv.FormatInt(value, 10)
	case uint64:
		stateValue = strconv.FormatUint(value, 10)
	case float32:
		stateValue = strconv.FormatFloat(float64(value), 'f', 3, 32)
	case float64:
		stateValue = strconv.FormatFloat(value, 'f', 3, 64)
	case bool:
		stateValue = strconv.FormatBool(value)
	default:
		stateValue = "unsupported"
	}
	if s.UnitOfMeasurement != "" {
		return stateValue + " " + s.UnitOfMeasurement
	}
	return stateValue
}
