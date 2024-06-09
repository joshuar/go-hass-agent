// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wsl
//revive:disable:unused-receiver
package sensor

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/registry"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

var mockSensor = SensorRegistrationMock{
	IDFunc:          func() string { return "mock_sensor" },
	StateFunc:       func() any { return "mockState" },
	AttributesFunc:  func() any { return nil },
	IconFunc:        func() string { return "mdi:mock-icon" },
	SensorTypeFunc:  func() types.SensorClass { return types.Sensor },
	NameFunc:        func() string { return "Mock Sensor" },
	UnitsFunc:       func() string { return "mockUnit" },
	StateClassFunc:  func() types.StateClass { return types.StateClassMeasurement },
	DeviceClassFunc: func() types.DeviceClass { return types.DeviceClassTemperature },
	CategoryFunc:    func() string { return "" },
}

// Code uses type checking against interfaces, so we need to define a mock type
// ourselves to ensure it explicitly satisfies the expected State interface.
type mockSensorType struct {
	value       any
	name, id    string
	sensorType  types.SensorClass
	deviceClass types.DeviceClass
	stateClass  types.StateClass
}

func (m *mockSensorType) Name() string                   { return m.name }
func (m *mockSensorType) ID() string                     { return m.id }
func (m *mockSensorType) State() any                     { return m.value }
func (m *mockSensorType) Icon() string                   { return "mdi:icon" }
func (m *mockSensorType) SensorType() types.SensorClass  { return m.sensorType }
func (m *mockSensorType) Units() string                  { return "" }
func (m *mockSensorType) Attributes() any                { return nil }
func (m *mockSensorType) DeviceClass() types.DeviceClass { return m.deviceClass }
func (m *mockSensorType) StateClass() types.StateClass   { return m.stateClass }
func (m *mockSensorType) Category() string               { return "" }

func Test_newStateUpdateRequest(t *testing.T) {
	type args struct {
		s State
	}
	tests := []struct {
		args args
		want *stateUpdateRequest
		name string
	}{
		{
			name: "default",
			args: args{s: &mockSensor},
			want: &stateUpdateRequest{
				UniqueID: "mock_sensor",
				Icon:     "mdi:mock-icon",
				Type:     types.Sensor.String(),
				State:    "mockState",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newStateUpdateRequest(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSensorState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newRegistrationRequest(t *testing.T) {
	type args struct {
		s Registration
	}
	tests := []struct {
		args args
		want *registrationRequest
		name string
	}{
		{
			name: "default",
			args: args{s: &mockSensor},
			want: &registrationRequest{
				stateUpdateRequest: &stateUpdateRequest{
					UniqueID: "mock_sensor",
					Icon:     "mdi:mock-icon",
					Type:     types.Sensor.String(),
					State:    "mockState",
				},
				Name:              "Mock Sensor",
				UnitOfMeasurement: "mockUnit",
				StateClass:        types.StateClassMeasurement.String(),
				DeviceClass:       types.DeviceClassTemperature.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newRegistrationRequest(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSensorRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:lll
func Test_request_RequestBody(t *testing.T) {
	var data []byte
	var err error
	data, err = json.Marshal(newStateUpdateRequest(&mockSensor))
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

//nolint:funlen,lll
//revive:disable:function-length
func TestNewRequest(t *testing.T) {
	registry.SetPath(filepath.Join(t.TempDir(), "testRegistry"))
	reg, err := registry.Load()
	require.NoError(t, err)
	err = reg.SetRegistered("sensor_update", true)
	require.NoError(t, err)
	err = reg.SetDisabled("sensor_disabled", true)
	require.NoError(t, err)

	mockLocation := &mockSensorType{
		name: "Location",
		id:   "location",
		value: &LocationRequest{
			Gps: []float64{0, 0},
		},
	}

	mockUpdate := &mockSensorType{
		name:  "Sensor Update",
		id:    "sensor_update",
		value: "value",
	}

	mockRegistration := &mockSensorType{
		name:  "Sensor Registration",
		id:    "sensor_registration",
		value: "value",
	}

	mockDisabled := &mockSensorType{
		name:  "Disabled Sensor",
		id:    "sensor_disabled",
		value: "value",
	}

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
			name: "location",
			args: args{reg: reg, req: mockLocation},
			want: &request{
				Data:        json.RawMessage(`{"gps":[0,0]}`),
				RequestType: requestTypeLocation,
			},
			want1:   &locationResponse{},
			wantErr: false,
		},
		{
			name: "update",
			args: args{reg: reg, req: mockUpdate},
			want: &request{
				Data:        json.RawMessage(`[{"state":"value","icon":"mdi:icon","type":"SensorClass(0)","unique_id":"sensor_update"}]`),
				RequestType: requestTypeUpdate,
			},
			want1:   &updateResponse{Body: make(map[string]*response)},
			wantErr: false,
		},
		{
			name: "registration",
			args: args{reg: reg, req: mockRegistration},
			want: &request{
				Data:        json.RawMessage(`{"state":"value","icon":"mdi:icon","type":"SensorClass(0)","unique_id":"sensor_registration","name":"Sensor Registration","state_class":"StateClass(0)","device_class":"DeviceClass(0)"}`),
				RequestType: requestTypeRegister,
			},
			want1:   &registrationResponse{},
			wantErr: false,
		},
		{
			name:    "disabled",
			args:    args{reg: reg, req: mockDisabled},
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
