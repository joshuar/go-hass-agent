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

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_sensorState_DeviceClass(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.SensorDeviceClass
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.DeviceClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.DeviceClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_StateClass(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.SensorStateClass
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.StateClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.StateClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_SensorType(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.SensorType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.SensorType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.SensorType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Icon(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.Icon(); got != tt.want {
				t.Errorf("sensorState.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Name(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("sensorState.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_State(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.State(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Attributes(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ID(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("sensorState.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Units(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.Units(); got != tt.want {
				t.Errorf("sensorState.Units() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Category(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:        tt.fields.data,
				disableCh:   tt.fields.disableCh,
				errCh:       tt.fields.errCh,
				requestData: tt.fields.requestData,
				requestType: tt.fields.requestType,
			}
			if got := s.Category(); got != tt.want {
				t.Errorf("sensorState.Category() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_RequestType(t *testing.T) {
	type fields struct {
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.RequestType
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
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
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
		data        hass.Sensor
		disableCh   chan bool
		errCh       chan error
		requestData []byte
		requestType hass.RequestType
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
	rSensor.On("SensorType").Return(hass.TypeSensor)
	rSensor.On("ID").Return("registeredID")
	rState := &sensorState{
		data:        rSensor,
		disableCh:   make(chan bool, 1),
		errCh:       make(chan error, 1),
		requestType: hass.RequestTypeUpdateSensorStates,
	}
	rState.requestData, err = json.Marshal(hass.MarshalSensorUpdate(rSensor))
	assert.Nil(t, err)

	uSensor := mocks.NewSensor(t)
	uSensor.On("Attributes").Return(nil)
	uSensor.On("DeviceClass").Return(hass.Duration)
	uSensor.On("Icon").Return("icon")
	uSensor.On("Name").Return("sensorName")
	uSensor.On("State").Return("state")
	uSensor.On("SensorType").Return(hass.TypeSensor)
	uSensor.On("ID").Return("unRegisteredID")
	uSensor.On("Units").Return("unit")
	uSensor.On("StateClass").Return(hass.StateMeasurement)
	uSensor.On("Category").Return("")
	uState := &sensorState{
		data:        uSensor,
		disableCh:   make(chan bool, 1),
		errCh:       make(chan error, 1),
		requestType: hass.RequestTypeRegisterSensor,
	}
	uState.requestData, err = json.Marshal(hass.MarshalSensorRegistration(uSensor))
	assert.Nil(t, err)

	r := NewMockRegistry(t)
	r.On("IsRegistered", "registeredID").Return(true)
	r.On("IsRegistered", "unRegisteredID").Return(false)

	type args struct {
		s hass.Sensor
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
