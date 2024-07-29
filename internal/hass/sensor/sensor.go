// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//nolint:errname // structs are dual-purpose response and error
//go:generate moq -out sensor_mocks_test.go . State Registration Details
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

var (
	ErrSensorDisabled  = errors.New("sensor disabled")
	ErrInvalidLocation = errors.New("invalid location update")
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

//nolint:exhaustruct // some fields are optional
func newStateUpdateRequest(sensor State) *stateUpdateRequest {
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

func generateRegistrationRequest(sensor Registration) *registrationRequest {
	reg := &registrationRequest{
		stateUpdateRequest: newStateUpdateRequest(sensor),
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
	RequestType string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}

func (r *Request) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}

	return json.RawMessage(data)
}

func NewUpdateRequest(sensor Details) (*Request, *StateUpdateResponse, error) {
	updates := []*stateUpdateRequest{newStateUpdateRequest(sensor)}

	data, err := json.Marshal(updates)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create state update request: %w", err)
	}

	return &Request{Data: data, RequestType: requestTypeUpdate},
		&StateUpdateResponse{Body: make(map[string]*UpdateStatus)},
		nil
}

func NewLocationUpdateRequest(sensor Details) (*Request, *LocationResponse, error) {
	location, ok := sensor.State().(*LocationRequest)
	if !ok {
		return nil, nil, ErrInvalidLocation
	}

	data, err := json.Marshal(location)
	if err != nil {
		return nil, nil, errors.Join(ErrInvalidLocation, err)
	}

	return &Request{Data: data, RequestType: requestTypeLocation},
		&LocationResponse{},
		nil
}

func NewRegistrationRequest(sensor Details) (*Request, *RegistrationResponse, error) {
	data, err := json.Marshal(generateRegistrationRequest(sensor))
	if err != nil {
		return nil, nil, fmt.Errorf("could not create registration request: %w", err)
	}

	return &Request{Data: data, RequestType: requestTypeRegister},
		&RegistrationResponse{},
		nil
}

type UpdateStatus struct {
	Error    *haError `json:"error,omitempty"`
	Success  bool     `json:"success,omitempty"`
	Disabled bool     `json:"is_disabled,omitempty"`
}

type haError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type StateUpdateResponse struct {
	Body           map[string]*UpdateStatus `json:"body"`
	*hass.APIError `json:"api_error,omitempty"`
}

func (u *StateUpdateResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &u.Body)
	if err != nil {
		return fmt.Errorf("could not parse response: %w", err)
	}

	return nil
}

func (u *StateUpdateResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, u.APIError)
	if err != nil {
		return fmt.Errorf("could not parse response error: %w", err)
	}

	return nil
}

func (u *StateUpdateResponse) Error() string {
	return u.APIError.Error()
}

func (u *StateUpdateResponse) Updates() map[string]*UpdateStatus {
	return u.Body
}

type RegistrationResponse struct {
	*hass.APIError
	Body UpdateStatus
}

func (r *RegistrationResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &r.Body)
	if err != nil {
		return fmt.Errorf("could not parse response: %w", err)
	}

	return nil
}

func (r *RegistrationResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, r.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (r *RegistrationResponse) Error() string {
	return r.APIError.Error()
}

func (r *RegistrationResponse) Registered() bool {
	return r.Body.Success
}

type LocationResponse struct {
	*hass.APIError
}

//revive:disable:unused-receiver
func (l *LocationResponse) UnmarshalJSON(_ []byte) error {
	return nil
}

func (l *LocationResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, l.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (l *LocationResponse) Error() string {
	return l.APIError.Error()
}

func (l *LocationResponse) Updated() bool {
	return l.APIError == nil
}
