// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl,nlreturn
//revive:disable:unused-parameter,function-length
//go:generate moq -out sensor_mocks_test.go . State Registration
package sensor

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

var (
	sensorExistingID  = "existing_sensor"
	sensorNewName     = "New Sensor"
	sensorNewID       = "new_sensor"
	sensorDisabledID  = "disabled_sensor"
	sensorIcon        = "mdi:icon"
	sensorType        = types.Sensor
	sensorState       = "sensorState"
	sensorUnits       = "sensorUnit"
	sensorDeviceClass = types.DeviceClassDataSize
	sensorStateClass  = types.StateClassMeasurement
)

var mockSensorState = &StateMock{
	AttributesFunc: func() map[string]any { return nil },
	IDFunc:         func() string { return sensorExistingID },
	IconFunc:       func() string { return sensorIcon },
	SensorTypeFunc: func() types.SensorClass { return sensorType },
	StateFunc:      func() any { return sensorState },
	UnitsFunc:      func() string { return sensorUnits },
}

var mockSensorRegistration = &RegistrationMock{
	AttributesFunc:  func() map[string]any { return nil },
	IDFunc:          func() string { return sensorNewID },
	IconFunc:        func() string { return sensorIcon },
	SensorTypeFunc:  func() types.SensorClass { return sensorType },
	StateFunc:       func() any { return sensorState },
	UnitsFunc:       func() string { return sensorUnits },
	CategoryFunc:    func() string { return "" },
	DeviceClassFunc: func() types.DeviceClass { return sensorDeviceClass },
	NameFunc:        func() string { return sensorNewName },
	StateClassFunc:  func() types.StateClass { return sensorStateClass },
}

var mockRegistry = &RegistryMock{
	IsRegisteredFunc: func(sensor string) bool {
		switch sensor {
		case sensorExistingID:
			return true
		case sensorNewID:
			return false
		}
		return false
	},
	IsDisabledFunc: func(sensor string) bool {
		return sensor == sensorDisabledID
	},
	SetRegisteredFunc: func(sensor string, state bool) error { return nil },
	SetDisabledFunc:   func(sensor string, state bool) error { return nil },
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
			args: args{sensor: mockSensorState},
			want: &stateUpdateRequest{
				State:    sensorState,
				Icon:     sensorIcon,
				Type:     sensorType.String(),
				UniqueID: sensorExistingID,
			},
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

func Test_newRegistrationRequest(t *testing.T) {
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
			args: args{sensor: mockSensorRegistration},
			want: &registrationRequest{
				stateUpdateRequest: &stateUpdateRequest{
					State:    sensorState,
					Icon:     sensorIcon,
					Type:     sensorType.String(),
					UniqueID: sensorNewID,
				},
				Name:        sensorNewName,
				DeviceClass: sensorDeviceClass.String(),
				StateClass:  sensorStateClass.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRegistrationRequest(tt.args.sensor)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.DeviceClass, got.DeviceClass)
			assert.Equal(t, tt.want.StateClass, got.StateClass)
		})
	}
}

func Test_request_RequestBody(t *testing.T) {
	sensor := newStateUpdateRequest(mockSensorState)
	validData, err := json.Marshal(sensor)
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
			fields: fields{RequestType: requestTypeUpdate, Data: validData},
			want:   json.RawMessage(`{"type":"update_sensor_states","data":` + string(validData) + `}`),
		},
		{
			name:   "invalid body",
			fields: fields{RequestType: requestTypeUpdate, Data: json.RawMessage(`invalid`)},
			want:   nil,
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

func TestNewRequest(t *testing.T) {
	newSensor := *mockSensorRegistration
	newSensorReq := newRegistrationRequest(&newSensor)
	newSensorData, err := json.Marshal(newSensorReq)
	require.NoError(t, err)
	newSensorRequest := &request{
		RequestType: requestTypeRegister,
		Data:        newSensorData,
	}

	existingSensor := *mockSensorRegistration
	existingSensor.IDFunc = func() string { return sensorExistingID }
	existingSensorReq := []*stateUpdateRequest{newStateUpdateRequest(&existingSensor)}
	existingSensorData, err := json.Marshal(existingSensorReq)
	require.NoError(t, err)
	existingSensorRequest := &request{
		RequestType: requestTypeUpdate,
		Data:        existingSensorData,
	}

	disabledSensor := *mockSensorRegistration
	disabledSensor.IDFunc = func() string { return sensorDisabledID }

	location := &LocationRequest{Gps: []float64{1.1, 2.2}}
	locationData, err := json.Marshal(location)
	require.NoError(t, err)
	locationDataRequest := &request{
		RequestType: requestTypeLocation,
		Data:        locationData,
	}
	locationSensor := *mockSensorRegistration
	locationSensor.StateFunc = func() any { return location }

	type args struct {
		reg Registry
		req any
	}
	tests := []struct {
		args    args
		want    hass.PostRequest
		want1   hass.Response
		name    string
		wantErr bool
	}{
		{
			name:  "new sensor",
			args:  args{reg: mockRegistry, req: Details(&newSensor)},
			want:  newSensorRequest,
			want1: &registrationResponse{},
		},
		{
			name:  "existing sensor",
			args:  args{reg: mockRegistry, req: Details(&existingSensor)},
			want:  existingSensorRequest,
			want1: &updateResponse{Body: make(map[string]*response)},
		},
		{
			name:  "location",
			args:  args{reg: mockRegistry, req: Details(&locationSensor)},
			want:  locationDataRequest,
			want1: &locationResponse{},
		},
		{
			name:    "disabled sensor",
			args:    args{reg: mockRegistry, req: Details(&disabledSensor)},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := NewRequest(tt.args.reg, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRequest() got = %v, want %v", string(got.RequestBody()), string(tt.want.RequestBody()))
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NewRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_updateResponse_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Body     map[string]*response
		APIError *hass.APIError
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid body",
			args: args{b: json.RawMessage(`{"sensor_id":{"success": true}}`)},
		},
		{
			name:    "empty body",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &updateResponse{
				Body:     tt.fields.Body,
				APIError: tt.fields.APIError,
			}
			if err := u.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("updateResponse.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_updateResponse_UnmarshalError(t *testing.T) {
	type fields struct {
		Body     map[string]*response
		APIError *hass.APIError
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid body",
			args:   args{data: json.RawMessage(`{"code":"501","message":"error"}`)},
			fields: fields{APIError: &hass.APIError{}},
		},
		{
			name:    "empty body",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &updateResponse{
				Body:     tt.fields.Body,
				APIError: tt.fields.APIError,
			}
			if err := u.UnmarshalError(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("updateResponse.UnmarshalError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_updateResponse_Error(t *testing.T) {
	type fields struct {
		Body     map[string]*response
		APIError *hass.APIError
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "with error",
			fields: fields{APIError: &hass.APIError{Code: "501", Message: "error"}},
			want:   "501: error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &updateResponse{
				Body:     tt.fields.Body,
				APIError: tt.fields.APIError,
			}
			if got := u.Error(); got != tt.want {
				t.Errorf("updateResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_registrationResponse_UnmarshalJSON(t *testing.T) {
	type fields struct {
		APIError *hass.APIError
		Body     response
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid body",
			args: args{b: json.RawMessage(`{"sensor_id":{"success": true}}`)},
		},
		{
			name:    "empty body",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registrationResponse{
				APIError: tt.fields.APIError,
				Body:     tt.fields.Body,
			}
			if err := r.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("registrationResponse.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_registrationResponse_UnmarshalError(t *testing.T) {
	type fields struct {
		APIError *hass.APIError
		Body     response
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid body",
			args:   args{data: json.RawMessage(`{"code":"501","message":"error"}`)},
			fields: fields{APIError: &hass.APIError{}},
		},
		{
			name:    "empty body",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registrationResponse{
				APIError: tt.fields.APIError,
				Body:     tt.fields.Body,
			}
			if err := r.UnmarshalError(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("registrationResponse.UnmarshalError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_registrationResponse_Error(t *testing.T) {
	type fields struct {
		APIError *hass.APIError
		Body     response
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "with error",
			fields: fields{APIError: &hass.APIError{Code: "501", Message: "error"}},
			want:   "501: error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registrationResponse{
				APIError: tt.fields.APIError,
				Body:     tt.fields.Body,
			}
			if got := r.Error(); got != tt.want {
				t.Errorf("registrationResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_locationResponse_UnmarshalJSON(t *testing.T) {
	type fields struct {
		APIError *hass.APIError
	}
	type args struct {
		in0 []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid body",
			args: args{in0: json.RawMessage(`{"sensor_id":{"success": true}}`)},
		},
		{
			name: "empty body",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &locationResponse{
				APIError: tt.fields.APIError,
			}
			if err := l.UnmarshalJSON(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("locationResponse.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_locationResponse_UnmarshalError(t *testing.T) {
	type fields struct {
		APIError *hass.APIError
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid body",
			args:   args{data: json.RawMessage(`{"code":"501","message":"error"}`)},
			fields: fields{APIError: &hass.APIError{}},
		},
		{
			name:    "empty body",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &locationResponse{
				APIError: tt.fields.APIError,
			}
			if err := l.UnmarshalError(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("locationResponse.UnmarshalError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_locationResponse_Error(t *testing.T) {
	type fields struct {
		APIError *hass.APIError
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "with error",
			fields: fields{APIError: &hass.APIError{Code: "501", Message: "error"}},
			want:   "501: error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &locationResponse{
				APIError: tt.fields.APIError,
			}
			if got := l.Error(); got != tt.want {
				t.Errorf("locationResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
