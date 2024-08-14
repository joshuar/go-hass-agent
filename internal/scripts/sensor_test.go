// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package scripts

import (
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

func TestScriptSensor_Name(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{SensorName: "script"},
			want:   "script",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("ScriptSensor.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_ID(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{SensorName: "Script"},
			want:   "script",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("scriptSensor.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_Icon(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "with icon",
			fields: fields{SensorIcon: "mdi:file"},
			want:   "mdi:file",
		},
		{
			name: "without icon",
			want: "mdi:script",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.Icon(); got != tt.want {
				t.Errorf("scriptSensor.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_SensorType(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   types.SensorClass
	}{
		{
			name:   "binary",
			fields: fields{SensorStateType: "binary"},
			want:   types.BinarySensor,
		},
		{
			name: "default",
			want: types.Sensor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.SensorType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scriptSensor.SensorType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_DeviceClass(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   types.DeviceClass
	}{
		{
			name:   "valid device class",
			fields: fields{SensorDeviceClass: types.DeviceClassDataRate.String()},
			want:   types.DeviceClassDataRate,
		},
		{
			name:   "invalid device class",
			fields: fields{SensorDeviceClass: "invalid"},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.DeviceClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scriptSensor.DeviceClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_StateClass(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   types.StateClass
	}{
		{
			name:   "valid state class",
			fields: fields{SensorStateClass: "measurement"},
			want:   types.StateClassMeasurement,
		},
		{
			name: "no state class",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.StateClass(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scriptSensor.StateClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_State(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		want   any
		fields fields
		name   string
	}{
		{
			name:   "default",
			fields: fields{SensorState: 1},
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.State(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scriptSensor.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_Units(t *testing.T) {
	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{SensorUnits: "%"},
			want:   "%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.Units(); got != tt.want {
				t.Errorf("scriptSensor.Units() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptSensor_Attributes(t *testing.T) {
	attrs := map[string]string{"attribute": "value"}

	type fields struct {
		SensorState       any
		SensorAttributes  any
		SensorName        string
		SensorIcon        string
		SensorDeviceClass string
		SensorStateClass  string
		SensorStateType   string
		SensorUnits       string
	}
	tests := []struct {
		want   map[string]any
		fields fields
		name   string
	}{
		{
			name:   "with attributes",
			fields: fields{SensorAttributes: attrs},
			want:   map[string]any{"extra_attributes": attrs},
		},
		{
			name: "without attributes",
			want: make(map[string]any),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScriptSensor{
				SensorState:       tt.fields.SensorState,
				SensorAttributes:  tt.fields.SensorAttributes,
				SensorName:        tt.fields.SensorName,
				SensorIcon:        tt.fields.SensorIcon,
				SensorDeviceClass: tt.fields.SensorDeviceClass,
				SensorStateClass:  tt.fields.SensorStateClass,
				SensorStateType:   tt.fields.SensorStateType,
				SensorUnits:       tt.fields.SensorUnits,
			}
			if got := s.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scriptSensor.Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}
