// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

func Test_sensorState_DeviceClass(t *testing.T) {
	type fields struct {
		state       interface{}
		attributes  interface{}
		metadata    *sensorMetadata
		stateUnits  string
		icon        string
		name        string
		entityID    string
		category    string
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.SensorDeviceClass
	}{
		{
			name:   "valid deviceClass",
			fields: fields{deviceClass: hass.Frequency},
			want:   hass.Frequency,
		},
		{
			name:   "unknown deviceClass",
			fields: fields{},
			want:   hass.SensorDeviceClass(0),
		},
		{
			name:   "invalid deviceClass",
			fields: fields{deviceClass: 65534},
			want:   hass.SensorDeviceClass(65534),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sensorState(tt.fields)
			if got := s.DeviceClass(); got != tt.want {
				t.Errorf("sensorState.DeviceClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_StateClass(t *testing.T) {
	type fields struct {
		state       interface{}
		attributes  interface{}
		metadata    *sensorMetadata
		stateUnits  string
		icon        string
		name        string
		entityID    string
		category    string
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.SensorStateClass
	}{
		{
			name:   "valid stateClass",
			fields: fields{stateClass: hass.StateMeasurement},
			want:   hass.StateMeasurement,
		},
		{
			name:   "no stateClass",
			fields: fields{},
			want:   hass.SensorStateClass(0),
		},
		{
			name:   "invalid deviceClass",
			fields: fields{stateClass: 65534},
			want:   hass.SensorStateClass(65534),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sensorState(tt.fields)

			if got := s.StateClass(); got != tt.want {
				t.Errorf("sensorState.StateClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Type(t *testing.T) {
	type fields struct {
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
		state       interface{}
		stateUnits  string
		attributes  interface{}
		icon        string
		name        string
		entityID    string
		category    string
		metadata    *sensorMetadata
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.SensorType
	}{
		{
			name:   "type sensor",
			fields: fields{sensorType: hass.TypeSensor},
			want:   hass.TypeSensor,
		},
		{
			name:   "type binary sensor",
			fields: fields{sensorType: hass.TypeBinary},
			want:   hass.TypeBinary,
		},
		{
			name:   "default sensor",
			fields: fields{},
			want:   hass.TypeSensor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				deviceClass: tt.fields.deviceClass,
				stateClass:  tt.fields.stateClass,
				sensorType:  tt.fields.sensorType,
				state:       tt.fields.state,
				stateUnits:  tt.fields.stateUnits,
				attributes:  tt.fields.attributes,
				icon:        tt.fields.icon,
				name:        tt.fields.name,
				entityID:    tt.fields.entityID,
				category:    tt.fields.category,
				metadata:    tt.fields.metadata,
			}
			if got := s.SensorType(); got != tt.want {
				t.Errorf("sensorState.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_State(t *testing.T) {
	type fields struct {
		state       interface{}
		attributes  interface{}
		metadata    *sensorMetadata
		stateUnits  string
		icon        string
		name        string
		entityID    string
		category    string
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name:   "valid state",
			fields: fields{state: "fakestate"},
			want:   "fakestate",
		},
		{
			name:   "default state",
			fields: fields{},
			want:   "Unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sensorState(tt.fields)
			if got := s.State(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_RequestType(t *testing.T) {
	type fields struct {
		state       interface{}
		attributes  interface{}
		metadata    *sensorMetadata
		stateUnits  string
		icon        string
		name        string
		entityID    string
		category    string
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := sensorState(tt.fields)
			if got := sensor.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_RequestData(t *testing.T) {
	registeredSensor := hass.MarshalSensorData(&sensorState{
		metadata: &sensorMetadata{Registered: true},
	})
	unRegisteredSensor := hass.MarshalSensorData(&sensorState{
		metadata: &sensorMetadata{Registered: false},
	})
	type fields struct {
		state       interface{}
		attributes  interface{}
		metadata    *sensorMetadata
		stateUnits  string
		icon        string
		name        string
		entityID    string
		category    string
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name:   "unregistered sensor",
			fields: fields{metadata: &sensorMetadata{Registered: false}},
			want:   unRegisteredSensor,
		},
		{
			name:   "registered sensor",
			fields: fields{metadata: &sensorMetadata{Registered: true}},
			want:   registeredSensor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := sensorState(tt.fields)
			if got := sensor.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ResponseHandler(t *testing.T) {
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
		deviceClass hass.SensorDeviceClass
		stateClass  hass.SensorStateClass
		sensorType  hass.SensorType
		state       interface{}
		stateUnits  string
		attributes  interface{}
		icon        string
		name        string
		entityID    string
		category    string
		metadata    *sensorMetadata
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
			name:   "registered sensor",
			fields: fields{},
			args:   args{rawResponse: *registeredResponse},
		},
		{
			name:   "updated sensor",
			fields: fields{},
			args:   args{rawResponse: *updatedResponse},
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
			name:   "no response",
			fields: fields{},
			args:   args{rawResponse: *bytes.NewBuffer(nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				deviceClass: tt.fields.deviceClass,
				stateClass:  tt.fields.stateClass,
				sensorType:  tt.fields.sensorType,
				state:       tt.fields.state,
				stateUnits:  tt.fields.stateUnits,
				attributes:  tt.fields.attributes,
				icon:        tt.fields.icon,
				name:        tt.fields.name,
				entityID:    tt.fields.entityID,
				category:    tt.fields.category,
				metadata:    tt.fields.metadata,
			}
			sensor.ResponseHandler(tt.args.rawResponse)
		})
	}
}

func Test_marshalSensorState(t *testing.T) {
	type args struct {
		s hass.SensorUpdate
	}
	tests := []struct {
		name string
		args args
		want *sensorState
	}{
		// TODO: add tests...
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := marshalSensorState(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalSensorState() = %v, want %v", got, tt.want)
			}
		})
	}
}
