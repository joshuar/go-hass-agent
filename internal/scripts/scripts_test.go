// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,lll,nlreturn,paralleltest,wsl,varnamelen,dupl
//revive:disable:unused-receiver
package scripts

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	jsonOut = `{"schedule":"@every 5s","sensors":[{"sensor_name": "random 1","sensor_icon": "mdi:dice-1","sensor_state":1},{"sensor_name": "random 2","sensor_icon": "mdi:dice-2","sensor_state_class":"measurement","sensor_state":6}]}`
	yamlOut = `schedule: '@every 5s'
sensors:
    - sensor_name: random 1
      sensor_icon: mdi:dice-1
      sensor_state: 8
    - sensor_name: random 2
      sensor_icon: mdi:dice-2
      sensor_state_class: measurement
      sensor_state: 9
`
	tomlOut = `schedule = '@every 5s'

[[sensors]]
sensor_icon = 'mdi:dice-1'
sensor_name = 'random 1'
sensor_state = 3

[[sensors]]
sensor_icon = 'mdi:dice-2'
sensor_name = 'random 2'
sensor_state = 3
sensor_state_class = 'measurement'
`
)

func TestScript_execute(t *testing.T) {
	var validOutput ScriptOutput
	err := json.Unmarshal([]byte(jsonOut), &validOutput)
	require.NoError(t, err)

	type fields struct {
		Path     string
		Schedule string
	}
	tests := []struct {
		want    *ScriptOutput
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "valid executable",
			fields: fields{Path: "/usr/bin/echo" + " " + jsonOut},
			want:   &validOutput,
		},
		{
			name:    "invalid executable",
			fields:  fields{Path: "/does/not/exist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Script{
				Path:     tt.fields.Path,
				Schedule: tt.fields.Schedule,
			}
			got, err := s.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Script.execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Script.execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptOutput_Unmarshal(t *testing.T) {
	type fields struct {
		Schedule string
		Sensors  []*scriptSensor
	}
	type args struct {
		scriptOutput []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "json output",
			args: args{scriptOutput: json.RawMessage(jsonOut)},
		},
		{
			name: "yaml output",
			args: args{scriptOutput: []byte(yamlOut)},
		},
		{
			name: "toml output",
			args: args{scriptOutput: []byte(tomlOut)},
		},
		{
			name:    "other output",
			args:    args{scriptOutput: []byte(`other`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &ScriptOutput{
				Schedule: tt.fields.Schedule,
				Sensors:  tt.fields.Sensors,
			}
			if err := o.Unmarshal(tt.args.scriptOutput); (err != nil) != tt.wantErr {
				t.Errorf("scriptOutput.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewScript(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		want    *Script
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid script",
			args: args{path: "/usr/bin/echo" + " " + jsonOut},
			want: &Script{
				Path:     "/usr/bin/echo" + " " + jsonOut,
				Schedule: "@every 5s",
			},
		},
		{
			name:    "invalid script",
			args:    args{path: "/does/not/exist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScript(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.Equal(t, tt.want.Path, got.Path)
				assert.Equal(t, tt.want.Schedule, got.Schedule)
			}
		})
	}
}

//nolint:containedctx
func TestFindScripts(t *testing.T) {
	script, err := NewScript("testing/data/jsonTestScript.sh")
	require.NoError(t, err)

	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Script
		wantErr bool
	}{
		{
			name: "path with scripts",
			args: args{ctx: context.TODO(), path: "testing/data"},
			want: []*Script{script},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindScripts(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got[0].Path, tt.want[0].Path) {
				t.Errorf("FindScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
			s := &scriptSensor{
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
