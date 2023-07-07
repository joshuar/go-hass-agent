// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
	"github.com/joshuar/go-hass-agent/internal/request"
	"github.com/joshuar/go-hass-agent/internal/tracker/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_sensorState_RequestType(t *testing.T) {
	type fields struct {
		data        Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType request.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   request.RequestType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := sensor.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_RequestData(t *testing.T) {
	type fields struct {
		data        Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType request.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := sensor.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ResponseHandler(t *testing.T) {
	rSensor := mocks.NewSensor(t)
	rSensor.On("ID").Return("registeredID")
	rSensor.On("Name").Return("sensorName")

	type fields struct {
		data        Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType request.RequestType
	}
	type args struct {
		rawResponse bytes.Buffer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "successful update",
			fields: fields{
				data:      rSensor,
				disableCh: make(chan bool, 1),
				errCh:     make(chan error, 1),
			},
			args: args{
				rawResponse: *bytes.NewBufferString(`{
					"registeredID": {
						"success": true
				}`),
			},
		},
		{
			name: "unsuccessful update ",
			fields: fields{
				data:      rSensor,
				disableCh: make(chan bool, 1),
				errCh:     make(chan error, 1),
			},
			args: args{
				rawResponse: *bytes.NewBufferString(`{
					"registeredID": {
						"success": false,
						"error": {
							"code": "invalid_format",
							"message": "Unexpected value for type"
						}
					}
				}`),
			},
		},
		{
			name: "empty response",
			fields: fields{
				data:      rSensor,
				disableCh: make(chan bool, 1),
				errCh:     make(chan error, 1),
			},
			args: args{rawResponse: *bytes.NewBuffer(nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			sensor.ResponseHandler(tt.args.rawResponse)
		})
	}
}

func Test_newSensorState(t *testing.T) {
	var err error

	rSensor := mocks.NewSensor(t)
	rSensor.On("Attributes").Return(nil)
	rSensor.On("Icon").Return("icon")
	rSensor.On("State").Return("state")
	rSensor.On("SensorType").Return(sensorType.TypeSensor)
	rSensor.On("ID").Return("registeredID")
	rState := &sensorState{
		data:        rSensor,
		disableCh:   make(chan bool, 1),
		errCh:       make(chan error, 1),
		requestType: request.RequestTypeUpdateSensorStates,
	}
	rState.requestData, err = json.Marshal(MarshalSensorUpdate(rSensor))
	assert.Nil(t, err)

	uSensor := mocks.NewSensor(t)
	uSensor.On("Attributes").Return(nil)
	uSensor.On("DeviceClass").Return(deviceClass.Duration)
	uSensor.On("Icon").Return("icon")
	uSensor.On("Name").Return("sensorName")
	uSensor.On("State").Return("state")
	uSensor.On("SensorType").Return(sensorType.TypeSensor)
	uSensor.On("ID").Return("unRegisteredID")
	uSensor.On("Units").Return("unit")
	uSensor.On("StateClass").Return(stateClass.StateMeasurement)
	uSensor.On("Category").Return("")
	uState := &sensorState{
		data:        uSensor,
		disableCh:   make(chan bool, 1),
		errCh:       make(chan error, 1),
		requestType: request.RequestTypeRegisterSensor,
	}
	uState.requestData, err = json.Marshal(MarshalSensorRegistration(uSensor))
	assert.Nil(t, err)

	r := mocks.NewRegistry(t)
	r.On("IsRegistered", "registeredID").Return(true)
	r.On("IsRegistered", "unRegisteredID").Return(false)

	type args struct {
		s Sensor
		r Registry
	}
	tests := []struct {
		name string
		args args
		want *sensorState
	}{
		{
			name: "registered",
			args: args{
				s: rSensor,
				r: r,
			},
			want: rState,
		},
		{
			name: "unregistered",
			args: args{
				s: uSensor,
				r: r,
			},
			want: uState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newSensorState(tt.args.s, tt.args.r)
			assert.Equal(t, got.data, tt.want.data)
			assert.Equal(t, got.requestData, tt.want.requestData)
			assert.Equal(t, got.requestType, tt.want.requestType)
		})
	}
}
