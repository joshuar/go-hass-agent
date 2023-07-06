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
	"github.com/stretchr/testify/mock"
)

func Test_sensorState_DeviceClass(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.DeviceClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.DeviceClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_StateClass(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.StateClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.StateClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_SensorType(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.SensorType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.SensorType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Icon(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Icon(); got != tt.want {
				t.Errorf("sensorState.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Name(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("sensorState.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_State(t *testing.T) {

	withState := mocks.NewSensorUpdate(t)
	withState.On("State", mock.AnythingOfType("string")).
		Return(func() interface{} {
			return "aString"
		})
	withoutState := mocks.NewSensorUpdate(t)
	withoutState.On("State").Return(nil)

	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "with state",
			fields: fields{
				data: withState,
			},
			want: "aString",
		},
		{
			name: "without state",
			fields: fields{
				data: withoutState,
			},
			want: "Unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.State(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Attributes(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ID(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("sensorState.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Units(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Units(); got != tt.want {
				t.Errorf("sensorState.Units() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Category(t *testing.T) {
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Category(); got != tt.want {
				t.Errorf("sensorState.Category() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Disabled(t *testing.T) {
	sensorUpdate := mocks.NewSensorUpdate(t)
	disabled := NewRegistryItem("sensorName")
	disabled.SetDisabled(true)

	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "disabled sensor",
			fields: fields{
				data:     sensorUpdate,
				metadata: disabled,
			},
			want: true,
		},
		{
			name: "active sensor",
			fields: fields{
				data:     sensorUpdate,
				metadata: NewRegistryItem("aSensor"),
			},
			want: false,
		},
		{
			name: "no metadata",
			fields: fields{
				data:     sensorUpdate,
				metadata: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Disabled(); got != tt.want {
				t.Errorf("sensorState.Disabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_Registered(t *testing.T) {
	sensorUpdate := mocks.NewSensorUpdate(t)
	registered := NewRegistryItem("aSensor")
	registered.SetRegistered(true)

	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "registered sensor",
			fields: fields{
				data:     sensorUpdate,
				metadata: registered,
			},
			want: true,
		},
		{
			name: "unregistered sensor",
			fields: fields{
				data:     sensorUpdate,
				metadata: NewRegistryItem("aSensor"),
			},
			want: false,
		},
		{
			name: "no metadata",
			fields: fields{
				data:     sensorUpdate,
				metadata: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := s.Registered(); got != tt.want {
				t.Errorf("sensorState.Registered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_MarshalJSON(t *testing.T) {
	aSensorState := mocks.NewSensorUpdate(t)
	aSensorState.On("Attributes").Return(nil)
	aSensorState.On("DeviceClass").Return(hass.Duration)
	aSensorState.On("Icon").Return("icon")
	aSensorState.On("Name").Return("sensorName")
	aSensorState.On("State").Return("state")
	aSensorState.On("SensorType").Return(hass.TypeSensor)
	aSensorState.On("ID").Return("sensorID")
	aSensorState.On("Units").Return("unit")
	aSensorState.On("StateClass").Return(hass.StateMeasurement)
	aSensorState.On("Category").Return("")

	registered := NewRegistryItem("sensorName")
	registered.SetRegistered(true)

	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "unregistered sensor",
			fields: fields{
				data: aSensorState,
			},
			want:    []byte(`{"device_class":"Duration","icon":"icon","name":"sensorName","state":"state","type":"sensor","unique_id":"sensorID","unit_of_measurement":"unit","state_class":"measurement"}`),
			wantErr: false,
		},
		{
			name: "registered sensor",
			fields: fields{
				data:     aSensorState,
				metadata: registered,
			},
			want:    []byte(`{"icon":"icon","state":"state","type":"sensor","unique_id":"sensorID"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
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
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
			name: "valid sensor data",
			args: args{
				b: []byte(`{"icon":"icon","state":"state","type":"sensor","unique_id":"sensorID"}`),
			},
		},
		{
			name: "invalid data",
			args: args{
				b: []byte(`{icon":"icon","state":"state","type":"sensor","unique_id":"sensorID"}`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if err := s.UnMarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("sensorState.UnMarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sensorState_RequestType(t *testing.T) {
	registered := NewRegistryItem("aSensor")
	registered.SetRegistered(true)

	unregistered := NewRegistryItem("aSensor")
	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   hass.RequestType
	}{
		{
			name:   "unregistered sensor",
			fields: fields{metadata: unregistered},
			want:   hass.RequestTypeRegisterSensor,
		},
		{
			name:   "registered sensor",
			fields: fields{metadata: registered},
			want:   hass.RequestTypeUpdateSensorStates,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := sensor.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_RequestData(t *testing.T) {
	aSensorState := mocks.NewSensorUpdate(t)
	aSensorState.On("Attributes").Return(nil)
	aSensorState.On("Icon").Return("icon")
	aSensorState.On("State").Return("state")
	aSensorState.On("SensorType").Return(hass.TypeSensor)
	aSensorState.On("ID").Return("sensorID")
	registered := NewRegistryItem("aSensor")
	registered.SetRegistered(true)

	type fields struct {
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		{
			name: "registered sensor",
			fields: fields{
				data:     aSensorState,
				metadata: registered,
			},
			want: json.RawMessage([]byte(`{"icon":"icon","state":"state","type":"sensor","unique_id":"sensorID"}`)),
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			if got := sensor.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorState.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorState_ResponseHandler(t *testing.T) {
	sensor := mocks.NewSensorUpdate(t)
	sensor.On("ID").Return("default")
	sensor.On("Name").Return("sensorName")
	registered := NewRegistryItem("aSensor")
	registered.SetRegistered(true)
	disabled := NewRegistryItem("aSensor")
	disabled.SetDisabled(true)

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
	disabledResponse := bytes.NewBufferString(`{
		"data": [
			{
		"battery_state": {
		  "success": true
		  "is_disabled": true
		}
	}
		],
	}`)
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
		data       hass.SensorUpdate
		metadata   *RegistryItem
		DisabledCh chan bool
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
				data:       sensor,
				metadata:   NewRegistryItem("sensorName"),
				DisabledCh: make(chan bool),
			},
			args: args{rawResponse: *registeredResponse},
		},
		{
			name: "updated sensor",
			fields: fields{
				data:       sensor,
				metadata:   registered,
				DisabledCh: make(chan bool),
			},
			args: args{rawResponse: *updatedResponse},
		},
		{
			name: "disabled sensor",
			fields: fields{
				data:       sensor,
				metadata:   NewRegistryItem("sensorName"),
				DisabledCh: make(chan bool),
			},
			args: args{rawResponse: *disabledResponse},
		},
		{
			name: "empty response",
			fields: fields{
				data:       sensor,
				metadata:   NewRegistryItem("sensorName"),
				DisabledCh: make(chan bool),
			},
			args: args{rawResponse: *bytes.NewBuffer(nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &sensorState{
				data:       tt.fields.data,
				metadata:   tt.fields.metadata,
				DisabledCh: tt.fields.DisabledCh,
			}
			sensor.ResponseHandler(tt.args.rawResponse)
		})
	}
}
