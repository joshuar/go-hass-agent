// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//go:generate moq -out sensor_mocks_test.go . State Registration Details
package sensor

import (
	"encoding/json"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	StateUnknown = "Unknown"

	CategoryDiagnostic = "diagnostic"

	requestTypeRegister = "register_sensor"
	requestTypeUpdate   = "update_sensor_states"
	requestTypeLocation = "update_location"
)

type State interface {
	ID() string
	Icon() string
	State() any
	SensorType() types.SensorClass
	Units() string
	Attributes() map[string]any
}

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

func createStateUpdateRequest(sensor State) *stateUpdateRequest {
	upd := &stateUpdateRequest{
		StateAttributes: sensor.Attributes(),
		State:           sensor.State(),
		Icon:            sensor.Icon(),
		UniqueID:        sensor.ID(),
	}

	if sensor.SensorType() > 0 {
		upd.Type = sensor.SensorType().String()
	}

	return upd
}

type registrationRequest struct {
	*stateUpdateRequest
	Name              string `json:"name,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	StateClass        string `json:"state_class,omitempty"`
	EntityCategory    string `json:"entity_category,omitempty"`
	DeviceClass       string `json:"device_class,omitempty"`
}

func createRegistrationRequest(sensor Registration) *registrationRequest {
	reg := &registrationRequest{
		stateUpdateRequest: createStateUpdateRequest(sensor),
		Name:               sensor.Name(),
		UnitOfMeasurement:  sensor.Units(),
		EntityCategory:     sensor.Category(),
	}

	if sensor.StateClass() > 0 {
		reg.StateClass = sensor.StateClass().String()
	}

	if sensor.DeviceClass() > 0 {
		reg.DeviceClass = sensor.DeviceClass().String()
	}

	return reg
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

type Request struct {
	Data        any    `json:"data"`
	RequestType string `json:"type"`
}

func (r *Request) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return json.RawMessage(data)
}

func NewRequest(details any) (*Request, error) {
	var (
		reqType string
		reqBody any
	)

	switch request := details.(type) {
	case Registration:
		reqBody = createRegistrationRequest(request)
		reqType = requestTypeRegister
	case State:
		reqBody = createStateUpdateRequest(request)
		reqType = requestTypeUpdate
	case LocationRequest:
		reqBody = request
		reqType = requestTypeLocation
	}

	return &Request{Data: reqBody, RequestType: reqType}, nil
}

type APIError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("code %s: %s", e.Code, e.Message)
}

type ResponseStatus struct {
	ErrorDetails *APIError
	IsSuccess    bool `json:"success,omitempty"`
}

type UpdateResponseStatus struct {
	ResponseStatus
	IsDisabled bool `json:"is_disabled,omitempty"`
}

func (u *UpdateResponseStatus) Disabled() bool {
	return u.IsDisabled
}

func (u *UpdateResponseStatus) Success() (bool, error) {
	if u.IsSuccess {
		return true, nil
	}

	return false, u.ErrorDetails
}

type StateUpdateResponse map[string]UpdateResponseStatus

type RegistrationResponse ResponseStatus

func (r *RegistrationResponse) Registered() (bool, error) {
	if r.IsSuccess {
		return true, nil
	}

	return false, r.ErrorDetails
}

type LocationResponse struct {
	error
}

//nolint:staticcheck
func (r *LocationResponse) Updated() error {
	return r
}
