// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/validation"
)

var (
	ErrMarshalSensor   = errors.New("could not marshal entity data")
	ErrUnmarshalSensor = errors.New("could not unmarshal entity data")
	ErrInvalidSensor   = errors.New("sensor data is invalid")
)

// Valid returns a boolean indicating whether the SensorState date is valid.
func (s *SensorState) Valid() (bool, error) {
	if err := validation.Validate.Struct(s); err != nil {
		return false, fmt.Errorf("%w: %s", ErrInvalidSensor, validation.ParseValidationErrors(err))
	}

	return true, nil
}

// Valid returns a boolean indicating whether the SensorRegistration data is valid.
func (s *SensorRegistration) Valid() (bool, error) {
	if err := validation.Validate.Struct(s); err != nil {
		return false, fmt.Errorf("%w: %s", ErrInvalidSensor, validation.ParseValidationErrors(err))
	}

	return true, nil
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
	if s.UnitOfMeasurement != nil {
		state = fmt.Sprintf("%v %s", s.State, *s.UnitOfMeasurement)
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
