// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package sensor

import (
	"encoding/json"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/validation"
)

const (
	StateUnknown = "Unknown"

	requestTypeRegisterSensor = "register_sensor"
	requestTypeUpdateSensor   = "update_sensor_states"
	requestTypeLocation       = "update_location"
)

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
