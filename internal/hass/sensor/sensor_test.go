// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockSensor = SensorRegistrationMock{
	IDFunc:          func() string { return "mock_sensor" },
	StateFunc:       func() any { return "mockState" },
	AttributesFunc:  func() any { return nil },
	IconFunc:        func() string { return "mdi:mock-icon" },
	SensorTypeFunc:  func() SensorType { return TypeSensor },
	NameFunc:        func() string { return "Mock Sensor" },
	UnitsFunc:       func() string { return "mockUnit" },
	StateClassFunc:  func() SensorStateClass { return StateMeasurement },
	DeviceClassFunc: func() SensorDeviceClass { return SensorTemperature },
	CategoryFunc:    func() string { return "" },
}

func Test_newSensorState(t *testing.T) {
	type args struct {
		s SensorState
	}
	tests := []struct {
		name string
		args args
		want *sensorState
	}{
		{
			name: "default",
			args: args{s: &mockSensor},
			want: &sensorState{
				UniqueID: "mock_sensor",
				Icon:     "mdi:mock-icon",
				Type:     marshalClass(TypeSensor),
				State:    "mockState",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSensorState(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSensorState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newSensorRegistration(t *testing.T) {
	type args struct {
		s SensorRegistration
	}
	tests := []struct {
		name string
		args args
		want *sensorRegistration
	}{
		{
			name: "default",
			args: args{s: &mockSensor},
			want: &sensorRegistration{
				sensorState: &sensorState{
					UniqueID: "mock_sensor",
					Icon:     "mdi:mock-icon",
					Type:     marshalClass(TypeSensor),
					State:    "mockState",
				},
				Name:              "Mock Sensor",
				UnitOfMeasurement: "mockUnit",
				StateClass:        marshalClass(StateMeasurement),
				DeviceClass:       marshalClass(SensorTemperature),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSensorRegistration(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSensorRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_request_RequestBody(t *testing.T) {
	var data []byte
	var err error
	data, err = json.Marshal(newSensorState(&mockSensor))
	assert.Nil(t, err)

	type fields struct {
		RequestType string
		Data        json.RawMessage
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		{
			name: "sensor state update",
			fields: fields{
				RequestType: requestTypeUpdate,
				Data:        data,
			},
			want: json.RawMessage(`{"type":"update_sensor_states","data":{"state":"mockState","icon":"mdi:mock-icon","type":"sensor","unique_id":"mock_sensor"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &request{
				RequestType: tt.fields.RequestType,
				Data:        tt.fields.Data,
			}
			if got := r.RequestBody(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("request.RequestBody() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func Test_request_ResponseBody(t *testing.T) {
	var stateData, registrationData []byte
	var err error
	stateData, err = json.Marshal(newSensorState(&mockSensor))
	assert.Nil(t, err)
	registrationData, err = json.Marshal(newSensorRegistration(&mockSensor))
	assert.Nil(t, err)

	type fields struct {
		RequestType string
		Data        json.RawMessage
	}
	tests := []struct {
		name   string
		fields fields
		want   any
	}{
		{
			name: "update",
			fields: fields{
				RequestType: requestTypeUpdate,
				Data:        stateData,
			},
			want: &UpdateResponse{},
		},
		{
			name: "registration",
			fields: fields{
				RequestType: requestTypeRegister,
				Data:        registrationData,
			},
			want: &RegistrationResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &request{
				RequestType: tt.fields.RequestType,
				Data:        tt.fields.Data,
			}
			if got := r.ResponseBody(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("request.ResponseBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateRequest(t *testing.T) {
	data, err := json.Marshal([]*sensorState{newSensorState(&mockSensor)})
	assert.Nil(t, err)

	type args struct {
		s []SensorState
	}
	tests := []struct {
		name string
		args args
		want *request
	}{
		{
			name: "default",
			args: args{s: []SensorState{&mockSensor}},
			want: &request{
				RequestType: requestTypeUpdate,
				Data:        data,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpdateRequest(tt.args.s...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistrationRequest(t *testing.T) {
	data, err := json.Marshal(newSensorRegistration(&mockSensor))
	assert.Nil(t, err)

	type args struct {
		s SensorRegistration
	}
	tests := []struct {
		name string
		args args
		want *request
	}{
		{
			name: "default",
			args: args{s: &mockSensor},
			want: &request{
				RequestType: requestTypeRegister,
				Data:        data,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RegistrationRequest(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegistrationRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_marshalClass(t *testing.T) {
	type args struct {
		class SensorDeviceClass
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "device class",
			args: args{class: SensorTemperature},
			want: "Temperature",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := marshalClass(tt.args.class); got != tt.want {
				t.Errorf("marshalClass() = %v, want %v", got, tt.want)
			}
		})
	}
}