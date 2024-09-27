// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package hass

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	requestTypeRegister = "register_sensor"
	requestTypeUpdate   = "update_sensor_states"
	requestTypeLocation = "update_location"
)

var (
	ErrNotLocation    = errors.New("sensor details do not represent a location update")
	ErrUnknownDetails = errors.New("unknown sensor details")
)

// LocationRequest represents the location information that can be sent to HA to
// update the location of the agent. This is exposed so that device code can
// create location requests directly, as Home Assistant handles these
// differently from other sensors.
type LocationRequest struct {
	Gps              []float64 `json:"gps"`
	GpsAccuracy      int       `json:"gps_accuracy,omitempty"`
	Battery          int       `json:"battery,omitempty"`
	Speed            int       `json:"speed,omitempty"`
	Altitude         int       `json:"altitude,omitempty"`
	Course           int       `json:"course,omitempty"`
	VerticalAccuracy int       `json:"vertical_accuracy,omitempty"`
}

type request struct {
	Data        any    `json:"data" validate:"required"`
	RequestType string `json:"type" validate:"required,oneof=register_sensor update_sensor_states update_location"`
}

func (r *request) Validate() error {
	err := validate.Struct(r)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrValidationFailed, parseValidationErrors(err))
	}

	return nil
}

func (r *request) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return json.RawMessage(data)
}

func newEntityRequest(requestType string, entity sensor.Entity) (*request, error) {
	switch requestType {
	case requestTypeLocation:
		return &request{Data: entity.State, RequestType: requestType}, nil
	case requestTypeRegister:
		return &request{Data: entity, RequestType: requestType}, nil
	case requestTypeUpdate:
		return &request{Data: entity.EntityState, RequestType: requestType}, nil
	}

	return nil, ErrUnknownDetails
}
