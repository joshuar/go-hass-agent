// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"encoding/json"
	"strings"
)

const (
	StateUnknown        = "unknown"
	requestTypeRegister = "register_sensor"
	requestTypeUpdate   = "update_sensor_states"
)

//go:generate moq -out mock_SensorState_test.go . SensorState
type SensorState interface {
	ID() string
	Icon() string
	State() any
	SensorType() SensorType
	Units() string
	Attributes() any
}

//go:generate moq -out mock_SensorRegistration_test.go . SensorRegistration
type SensorRegistration interface {
	SensorState
	Name() string
	DeviceClass() SensorDeviceClass
	StateClass() SensorStateClass
	Category() string
}

type Details interface {
	SensorState
	SensorRegistration
}

type sensorState struct {
	StateAttributes any    `json:"attributes,omitempty"`
	State           any    `json:"state"`
	Icon            string `json:"icon,omitempty"`
	Type            string `json:"type"`
	UniqueID        string `json:"unique_id"`
}

func newSensorState(s SensorState) *sensorState {
	return &sensorState{
		StateAttributes: s.Attributes(),
		State:           s.State(),
		Icon:            s.Icon(),
		Type:            marshalClass(s.SensorType()),
		UniqueID:        s.ID(),
	}
}

type sensorRegistration struct {
	*sensorState
	Name              string `json:"name,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	StateClass        string `json:"state_class,omitempty"`
	EntityCategory    string `json:"entity_category,omitempty"`
	DeviceClass       string `json:"device_class,omitempty"`
}

func newSensorRegistration(s SensorRegistration) *sensorRegistration {
	return &sensorRegistration{
		sensorState:       newSensorState(s),
		Name:              s.Name(),
		UnitOfMeasurement: s.Units(),
		StateClass:        marshalClass(s.StateClass()),
		EntityCategory:    s.Category(),
		DeviceClass:       marshalClass(s.DeviceClass()),
	}
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

func (r *request) ResponseBody() any {
	if strings.Contains(r.RequestType, requestTypeUpdate) {
		resp := make(UpdateResponse)
		return &resp
	}
	if strings.Contains(r.RequestType, requestTypeRegister) {
		return &RegistrationResponse{}
	}
	resp := make(map[string]any)
	return &resp
}

func UpdateRequest(s ...SensorState) *request {
	var updates []*sensorState
	for _, u := range s {
		updates = append(updates, newSensorState(u))
	}
	data, err := json.Marshal(updates)
	if err != nil {
		return nil
	}
	return &request{
		Data:        data,
		RequestType: requestTypeUpdate,
	}
}

func RegistrationRequest(s SensorRegistration) *request {
	data, err := json.Marshal(newSensorRegistration(s))
	if err != nil {
		return nil
	}
	return &request{
		Data:        data,
		RequestType: requestTypeRegister,
	}
}

type response struct {
	Error    *details `json:"error,omitempty"`
	Success  bool     `json:"success,omitempty"`
	Disabled bool     `json:"is_disabled,omitempty"`
}

type details struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type UpdateResponse map[string]*response

type RegistrationResponse response

type comparableStringer interface {
	comparable
	String() string
}

func returnZero[T any](s ...T) T {
	var zero T
	return zero
}

func marshalClass[C comparableStringer](class C) string {
	if class == returnZero[C](class) {
		return ""
	} else {
		return class.String()
	}
}
