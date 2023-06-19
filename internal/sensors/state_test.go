// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/mock"
)

type mockSensorUpdate struct {
	mock.Mock
}

func (m *mockSensorUpdate) Attributes() interface{} {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) DeviceClass() hass.SensorDeviceClass {
	args := m.Called()
	return args.Get(0).(hass.SensorDeviceClass)
}

func (m *mockSensorUpdate) Icon() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) State() interface{} {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) SensorType() hass.SensorType {
	args := m.Called()
	return args.Get(0).(hass.SensorType)
}

func (m *mockSensorUpdate) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) Units() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) StateClass() hass.SensorStateClass {
	args := m.Called()
	return args.Get(0).(hass.SensorStateClass)
}

func (m *mockSensorUpdate) Category() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) Registered() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSensorUpdate) Disabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSensorUpdate) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockSensorUpdate) UnMarshalJSON(b []byte) error {
	args := m.Called(b)
	return args.Error(1)
}

func Test_sensorState_DeviceClass(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.DeviceClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.DeviceClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_StateClass(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.StateClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.StateClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_SensorType(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.SensorType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.SensorType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Icon(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Icon(); got != tt.want {
				t.Errorf("sensorState.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Name(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("sensorState.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_State(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.State(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Attributes(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ID(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("sensorState.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Units(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Units(); got != tt.want {
				t.Errorf("sensorState.Units() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Category(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Category(); got != tt.want {
				t.Errorf("sensorState.Category() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Disabled(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Disabled(); got != tt.want {
				t.Errorf("sensorState.Disabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Registered(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := s.Registered(); got != tt.want {
				t.Errorf("sensorState.Registered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_MarshalJSON(t *testing.T) {
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "test registered",
			fields: fields{
				metadata: &sensorMetadata{
					Registered: true,
				},
				data: s,
			},
			want:    []byte(`{"attributes":"","icon":"default","state":"default","type":"sensor","unique_id":"default"}`),
			wantErr: false,
		},
		{
			name: "test unregistered",
			fields: fields{
				metadata: &sensorMetadata{
					Registered: false,
				},
				data: s,
			},
			want:    []byte(`{"attributes":"","device_class":"Duration","icon":"default","name":"default","state":"default","type":"sensor","unique_id":"default","state_class":"measurement"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			got, err := s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("sensorState.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_UnMarshalJSON(t *testing.T) {
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
			name: "default test",
			fields: fields{
				data: s,
			},
			args: args{
				b: []byte(`{"name":"default","state":"default","type":"sensor"}`),
			},
			wantErr: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if err := s.UnMarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("sensorState.UnMarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sensorState_RequestType(t *testing.T) {
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.RequestType
	}{
		{
			name:   "unregistered sensor",
			fields: fields{metadata: &sensorMetadata{Registered: false}},
			want:   hass.RequestTypeRegisterSensor,
		},
		{
			name:   "registered sensor",
			fields: fields{metadata: &sensorMetadata{Registered: true}},
			want:   hass.RequestTypeUpdateSensorStates,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := sensor.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_RequestData(t *testing.T) {
	defaultMsg := json.RawMessage(`{"attributes":"","device_class":"Duration","icon":"default","name":"default","state":"default","type":"sensor","unique_id":"default","state_class":"measurement"}`)
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
	}
	tests := []struct {
		name   string
		fields fields
		want   *json.RawMessage
	}{
		{
			name: "default test",
			fields: fields{
				data:     s,
				metadata: &sensorMetadata{},
			},
			want: &defaultMsg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			if got := sensor.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ResponseHandler(t *testing.T) {
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")
	registeredResponse := bytes.NewBufferString(`{
		"data": {
		  "attributes": {
			"foo": "bar"
		  },
		  "device_class": "battery",
		  "icon": "mdi:battery",
		  "name": "Battery State",
		  "state": "12345",
		  "type": "sensor",
		  "unique_id": "battery_state",
		  "unit_of_measurement": "%",
		  "state_class": "measurement",
		  "entity_category": "diagnostic",
		  "disabled": true
		},
		"type": "register_sensor"
	  }`)
	updatedResponse := bytes.NewBufferString(`{
		"data": [
		  {
			"attributes": {
			  "hello": "world"
			},
			"icon": "mdi:battery",
			"state": 123,
			"type": "sensor",
			"unique_id": "battery_state"
		  }
		],
		"type": "update_sensor_states"
	  }`)
	// TODO: fix the format of these responses...
	// disabledResponse := bytes.NewBufferString(`{
	// 	"data": [
	// 		{
	// 	"battery_state": {
	// 	  "success": true
	// 	  "is_disabled": true
	// 	}
	// }
	// 	],
	// }`)
	// unRegisteredResponse := bytes.NewBufferString(`{
	// 	"battery_charging": {
	// 	  "success": false,
	// 	  "error": {
	// 		"code": "not_registered",
	// 		"message": "Entity is not registered",
	// 	  }
	// 	}
	// }`)
	// errorResponse := bytes.NewBufferString(`{
	// 	"battery_charging_state": {
	// 	  "success": false,
	// 	  "error": {
	// 		"code": "invalid_format",
	// 		"message": "Unexpected value for type",
	// 	  }
	//   }`)
	type fields struct {
		data     hass.SensorUpdate
		metadata *sensorMetadata
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
			name: "registered sensor",
			fields: fields{
				data: s,
				metadata: &sensorMetadata{
					Registered: true,
				},
			},
			args: args{rawResponse: *registeredResponse},
		},
		{
			name: "updated sensor",
			fields: fields{
				data: s,
				metadata: &sensorMetadata{
					Registered: true,
				},
			},
			args: args{rawResponse: *updatedResponse},
		},
		// {
		// 	name:   "disabled sensor",
		// 	fields: fields{},
		// 	args:   args{rawResponse: *disabledResponse},
		// },
		// {
		// 	name:   "unregistered sensor",
		// 	fields: fields{},
		// 	args:   args{rawResponse: *unRegisteredResponse},
		// },
		// {
		// 	name:   "error response",
		// 	fields: fields{},
		// 	args:   args{rawResponse: *errorResponse},
		// },
		{
			name: "no response",
			fields: fields{
				data:     s,
				metadata: &sensorMetadata{},
			},
			args: args{rawResponse: *bytes.NewBuffer(nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:     tt.fields.data,
				metadata: tt.fields.metadata,
			}
			sensor.ResponseHandler(tt.args.rawResponse)
		})
	}
}
