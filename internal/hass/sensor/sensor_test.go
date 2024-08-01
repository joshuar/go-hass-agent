// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//nolint:paralleltest,wsl,dupl,nlreturn
//go:generate moq -out sensor_mocks_test.go . State Registration Details
package sensor

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	mockUpdateValue = "someValue"
	mockUpdateID    = "someID"
	mockIcon        = "mdi:icon"
	mockSensorType  = types.Sensor
	mockRegValue    = "someValue"
	mockRegID       = "someID"
	mockDeviceClass = types.DeviceClassDataSize
	mockStateClass  = types.StateClassMeasurement
	mockUnits       = "KB"
	mockName        = "Sensor"
)

var (
	mockUpdate = &StateMock{
		StateFunc:      func() any { return mockUpdateValue },
		IconFunc:       func() string { return mockIcon },
		IDFunc:         func() string { return mockUpdateID },
		SensorTypeFunc: func() types.SensorClass { return mockSensorType },
		AttributesFunc: func() map[string]any { return nil },
	}
	mockUpdateRequest = &stateUpdateRequest{
		State:    mockRegValue,
		Icon:     mockIcon,
		Type:     mockSensorType.String(),
		UniqueID: mockRegID,
	}
)

var (
	mockRegistration = &RegistrationMock{
		StateFunc:       func() any { return mockRegValue },
		IconFunc:        func() string { return mockIcon },
		IDFunc:          func() string { return mockRegID },
		SensorTypeFunc:  func() types.SensorClass { return mockSensorType },
		AttributesFunc:  func() map[string]any { return nil },
		CategoryFunc:    func() string { return "" },
		DeviceClassFunc: func() types.DeviceClass { return mockDeviceClass },
		StateClassFunc:  func() types.StateClass { return mockStateClass },
		UnitsFunc:       func() string { return mockUnits },
		NameFunc:        func() string { return mockName },
	}
	mockRegistrationRequest = &registrationRequest{
		stateUpdateRequest: mockUpdateRequest,
		Name:               mockName,
		UnitOfMeasurement:  mockUnits,
		DeviceClass:        mockDeviceClass.String(),
		StateClass:         mockStateClass.String(),
		EntityCategory:     "",
	}
)

var mockDetails = &DetailsMock{
	StateFunc:       func() any { return mockRegValue },
	IconFunc:        func() string { return mockIcon },
	IDFunc:          func() string { return mockRegID },
	SensorTypeFunc:  func() types.SensorClass { return mockSensorType },
	AttributesFunc:  func() map[string]any { return nil },
	CategoryFunc:    func() string { return "" },
	DeviceClassFunc: func() types.DeviceClass { return mockDeviceClass },
	StateClassFunc:  func() types.StateClass { return mockStateClass },
	UnitsFunc:       func() string { return mockUnits },
	NameFunc:        func() string { return mockName },
}

func Test_newStateUpdateRequest(t *testing.T) {
	type args struct {
		sensor State
	}
	tests := []struct {
		args args
		want *stateUpdateRequest
		name string
	}{
		{
			name: "default",
			args: args{sensor: mockUpdate},
			want: mockUpdateRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newStateUpdateRequest(tt.args.sensor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newStateUpdateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateRegistrationRequest(t *testing.T) {
	type args struct {
		sensor Registration
	}
	tests := []struct {
		args args
		want *registrationRequest
		name string
	}{
		{
			name: "default",
			args: args{sensor: mockRegistration},
			want: mockRegistrationRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateRegistrationRequest(tt.args.sensor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateRegistrationRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_RequestBody(t *testing.T) {
	data, err := json.Marshal(mockUpdateRequest)
	require.NoError(t, err)

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
			name:   "valid body",
			fields: fields{RequestType: requestTypeUpdate, Data: data},
			want:   json.RawMessage(`{"type":"` + requestTypeUpdate + `","data":` + string(data) + `}`),
		},
		{
			name:   "invalid body",
			fields: fields{RequestType: requestTypeUpdate, Data: []byte(`foo`)},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Request{
				RequestType: tt.fields.RequestType,
				Data:        tt.fields.Data,
			}
			if got := r.RequestBody(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Request.RequestBody() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestNewUpdateRequest(t *testing.T) {
	data, err := json.Marshal([]*stateUpdateRequest{mockUpdateRequest})
	require.NoError(t, err)

	type args struct {
		sensor Details
	}
	tests := []struct {
		args    args
		want    *Request
		want1   *StateUpdateResponse
		name    string
		wantErr bool
	}{
		{
			name:  "valid update",
			args:  args{sensor: mockDetails},
			want:  &Request{RequestType: requestTypeUpdate, Data: data},
			want1: &StateUpdateResponse{Body: make(map[string]*UpdateStatus)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := NewUpdateRequest(tt.args.sensor)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUpdateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUpdateRequest() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewUpdateRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewLocationUpdateRequest(t *testing.T) {
	mockLocation := *mockDetails

	mockLocation.StateFunc = func() any {
		return &LocationRequest{
			Gps: []float64{1, 1},
		}
	}

	data, err := json.Marshal(&LocationRequest{
		Gps: []float64{1, 1},
	})
	require.NoError(t, err)

	type args struct {
		sensor Details
	}
	tests := []struct {
		args    args
		want    *Request
		want1   *LocationResponse
		name    string
		wantErr bool
	}{
		{
			name:  "valid location",
			args:  args{sensor: &mockLocation},
			want:  &Request{RequestType: requestTypeLocation, Data: data},
			want1: &LocationResponse{},
		},
		{
			name:    "invalid location",
			args:    args{sensor: mockDetails},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := NewLocationUpdateRequest(tt.args.sensor)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocationUpdateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLocationUpdateRequest() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewLocationUpdateRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewRegistrationRequest(t *testing.T) {
	data, err := json.Marshal(mockRegistrationRequest)
	require.NoError(t, err)

	type args struct {
		sensor Details
	}
	tests := []struct {
		args    args
		want    *Request
		want1   *RegistrationResponse
		name    string
		wantErr bool
	}{
		{
			name:  "valid registration",
			args:  args{sensor: mockDetails},
			want:  &Request{RequestType: requestTypeRegister, Data: data},
			want1: &RegistrationResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := NewRegistrationRequest(tt.args.sensor)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegistrationRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRegistrationRequest() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewRegistrationRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
