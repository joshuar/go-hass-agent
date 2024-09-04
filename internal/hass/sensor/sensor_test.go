// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

func newMockUpdate(t *testing.T) (State, *stateUpdateRequest) {
	t.Helper()
	state := "updateState"
	icon := "mdi:update"
	id := "updateID"

	return &StateMock{
			StateFunc:      func() any { return state },
			IconFunc:       func() string { return icon },
			IDFunc:         func() string { return id },
			SensorTypeFunc: func() types.SensorClass { return types.Sensor },
			AttributesFunc: func() map[string]any { return nil },
		},
		&stateUpdateRequest{
			State:    state,
			Icon:     icon,
			UniqueID: id,
			Type:     types.Sensor.String(),
		}
}

func newMockRegistration(t *testing.T) (Registration, *registrationRequest) {
	t.Helper()
	state := "regState"
	icon := "mdi:registration"
	id := "regID"
	name := "registration"
	units := "units"

	return &RegistrationMock{
			StateFunc:       func() any { return state },
			IconFunc:        func() string { return icon },
			IDFunc:          func() string { return id },
			NameFunc:        func() string { return name },
			UnitsFunc:       func() string { return units },
			SensorTypeFunc:  func() types.SensorClass { return types.Sensor },
			AttributesFunc:  func() map[string]any { return nil },
			CategoryFunc:    func() string { return "" },
			StateClassFunc:  func() types.StateClass { return types.StateClassMeasurement },
			DeviceClassFunc: func() types.DeviceClass { return types.DeviceClassDataRate },
		},
		&registrationRequest{
			stateUpdateRequest: &stateUpdateRequest{
				State:    state,
				Icon:     icon,
				UniqueID: id,
				Type:     types.Sensor.String(),
			},
			Name:              name,
			UnitOfMeasurement: units,
			StateClass:        types.StateClassMeasurement.String(),
			DeviceClass:       types.DeviceClassDataRate.String(),
		}
}

func newMockDetails(t *testing.T) (Details, *registrationRequest, *stateUpdateRequest) {
	t.Helper()
	state := "regState"
	icon := "mdi:registration"
	id := "regID"
	name := "registration"
	units := "units"
	updateReq := &stateUpdateRequest{
		State:    state,
		Icon:     icon,
		UniqueID: id,
		Type:     types.Sensor.String(),
	}
	regReq := &registrationRequest{
		stateUpdateRequest: updateReq,
		Name:               name,
		UnitOfMeasurement:  units,
		StateClass:         types.StateClassMeasurement.String(),
		DeviceClass:        types.DeviceClassDataRate.String(),
	}
	return &DetailsMock{
			StateFunc:       func() any { return state },
			IconFunc:        func() string { return icon },
			IDFunc:          func() string { return id },
			NameFunc:        func() string { return name },
			UnitsFunc:       func() string { return units },
			SensorTypeFunc:  func() types.SensorClass { return types.Sensor },
			AttributesFunc:  func() map[string]any { return nil },
			CategoryFunc:    func() string { return "" },
			StateClassFunc:  func() types.StateClass { return types.StateClassMeasurement },
			DeviceClassFunc: func() types.DeviceClass { return types.DeviceClassDataRate },
		},
		regReq,
		updateReq
}

func newMockLocation(t *testing.T) (Details, *LocationRequest) {
	t.Helper()
	return &DetailsMock{
			StateFunc: func() any {
				return &LocationRequest{
					Gps: []float64{33, 34},
				}
			},
		},
		&LocationRequest{
			Gps: []float64{33, 34},
		}
}

func Test_createStateUpdateRequest(t *testing.T) {
	validState, validStateReq := newMockUpdate(t)
	type args struct {
		sensor State
	}
	tests := []struct {
		args args
		want *stateUpdateRequest
		name string
	}{
		{
			name: "valid update",
			args: args{sensor: validState},
			want: validStateReq,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createStateUpdateRequest(tt.args.sensor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createStateUpdateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createRegistrationRequest(t *testing.T) {
	validReg, validRegReq := newMockRegistration(t)

	type args struct {
		sensor Registration
	}
	tests := []struct {
		args args
		want *registrationRequest
		name string
	}{
		{
			name: "valid registration",
			args: args{sensor: validReg},
			want: validRegReq,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createRegistrationRequest(tt.args.sensor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createRegistrationRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_RequestBody(t *testing.T) {
	ja := jsonassert.New(t)

	_, validStateReq := newMockUpdate(t)
	data, err := json.Marshal(validStateReq)
	require.NoError(t, err)

	type fields struct {
		Data        any
		RequestType string
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		{
			name:   "valid body",
			fields: fields{RequestType: RequestTypeUpdate, Data: validStateReq},
			want:   json.RawMessage(`{"type":"` + RequestTypeUpdate + `","data":` + string(data) + `}`),
		},
		{
			name:   "invalid body",
			fields: fields{RequestType: RequestTypeUpdate, Data: []byte(`foo`)},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { //revive:disable:unused-parameter
			r := &Request{
				Data:        tt.fields.Data,
				RequestType: tt.fields.RequestType,
			}
			got := r.RequestBody()
			if tt.want != nil {
				ja.Assertf(string(got), string(tt.want)) //nolint:govet
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	details, regReq, updReq := newMockDetails(t)
	loc, locReq := newMockLocation(t)

	type args struct {
		details Details
		reqType string
	}
	tests := []struct {
		want    *Request
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "state update",
			args: args{details: details, reqType: RequestTypeUpdate},
			want: &Request{RequestType: RequestTypeUpdate, Data: updReq},
		},
		{
			name: "registration",
			args: args{details: details, reqType: RequestTypeRegister},
			want: &Request{RequestType: RequestTypeRegister, Data: regReq},
		},
		{
			name: "location update",
			args: args{details: loc, reqType: RequestTypeLocation},
			want: &Request{RequestType: RequestTypeLocation, Data: locReq},
		},
		{
			name:    "unknown",
			args:    args{details: details, reqType: "unknown"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRequest(tt.args.reqType, tt.args.details)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
