// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	StateUnknown = "Unknown"

	requestTypeRegister = "register_sensor"
	requestTypeUpdate   = "update_sensor_states"
	requestTypeLocation = "update_location"
)

var ErrSensorDisabled = errors.New("sensor disabled")

//go:generate moq -out mock_State_test.go . State
type State interface {
	ID() string
	Icon() string
	State() any
	SensorType() types.SensorClass
	Units() string
	Attributes() map[string]any
}

//go:generate moq -out mock_Registration_test.go . Registration
type Registration interface {
	State
	Name() string
	DeviceClass() types.DeviceClass
	StateClass() types.StateClass
	Category() string
}

type Details interface {
	State
	Registration
}

type stateUpdateRequest struct {
	StateAttributes map[string]any `json:"attributes,omitempty"`
	State           any            `json:"state"`
	Icon            string         `json:"icon,omitempty"`
	Type            string         `json:"type"`
	UniqueID        string         `json:"unique_id"`
}

func newStateUpdateRequest(sensor State) *stateUpdateRequest {
	return &stateUpdateRequest{
		StateAttributes: sensor.Attributes(),
		State:           sensor.State(),
		Icon:            sensor.Icon(),
		Type:            sensor.SensorType().String(),
		UniqueID:        sensor.ID(),
	}
}

type registrationRequest struct {
	*stateUpdateRequest
	Name              string `json:"name,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	StateClass        string `json:"state_class,omitempty"`
	EntityCategory    string `json:"entity_category,omitempty"`
	DeviceClass       string `json:"device_class,omitempty"`
}

func newRegistrationRequest(sensor Registration) *registrationRequest {
	return &registrationRequest{
		stateUpdateRequest: newStateUpdateRequest(sensor),
		Name:               sensor.Name(),
		UnitOfMeasurement:  sensor.Units(),
		StateClass:         sensor.StateClass().String(),
		EntityCategory:     sensor.Category(),
		DeviceClass:        sensor.DeviceClass().String(),
	}
}

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
	RequestType string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}

func (r *request) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return json.RawMessage(data)
}

//nolint:exhaustruct,err113
//revive:disable:unnecessary-stmt
func NewRequest(reg Registry, req any) (hass.PostRequest, hass.Response, error) {
	switch sensor := req.(type) {
	case Details:
		// Location Request is a special case.
		if location, ok := sensor.State().(*LocationRequest); ok {
			data, err := json.Marshal(location)
			if err != nil {
				return nil, nil, fmt.Errorf("could not create location request: %w", err)
			}

			return &request{Data: data, RequestType: requestTypeLocation},
				&locationResponse{},
				nil
		}
		// If the sensor is disabled, don't bother creating a request.
		if reg.IsDisabled(sensor.ID()) {
			return nil, nil, ErrSensorDisabled
		}

		if reg.IsRegistered(sensor.ID()) {
			// If the sensor is registered, create an update request.
			updates := []*stateUpdateRequest{newStateUpdateRequest(sensor)}

			data, err := json.Marshal(updates)
			if err != nil {
				return nil, nil, fmt.Errorf("could not create state update request: %w", err)
			}

			return &request{Data: data, RequestType: requestTypeUpdate},
				&updateResponse{Body: make(map[string]*response)},
				nil
		} else {
			// Else, create a registration request.
			data, err := json.Marshal(newRegistrationRequest(sensor))
			if err != nil {
				return nil, nil, fmt.Errorf("could not create registration request: %w", err)
			}

			return &request{Data: data, RequestType: requestTypeRegister},
				&registrationResponse{},
				nil
		}
	}

	return nil, nil, fmt.Errorf("unknown request type: %T", req)
}

type response struct {
	Error    *haError `json:"error,omitempty"`
	Success  bool     `json:"success,omitempty"`
	Disabled bool     `json:"is_disabled,omitempty"`
}

type haError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type updateResponse struct {
	Body map[string]*response
}

func (u *updateResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &u.Body)
	if err != nil {
		return fmt.Errorf("could not parse response: %w", err)
	}

	return nil
}

type registrationResponse struct {
	Body response
}

func (r *registrationResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &r.Body)
	if err != nil {
		return fmt.Errorf("could not parse response: %w", err)
	}

	return nil
}

type locationResponse struct{}

//revive:disable:unused-receiver
func (l *locationResponse) UnmarshalJSON(_ []byte) error {
	return nil
}
