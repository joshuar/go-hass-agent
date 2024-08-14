// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
//revive:disable:unused-receiver
package scripts

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	//nolint:lll
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

func Test_scriptOutput_Unmarshal(t *testing.T) {
	type fields struct {
		Schedule string
		Sensors  []ScriptSensor
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
			o := &scriptOutput{
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
				path:     "/usr/bin/echo" + " " + jsonOut,
				schedule: "@every 5s",
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
				assert.Equal(t, tt.want.path, got.path)
				assert.Equal(t, tt.want.Schedule(), got.Schedule())
			}
		})
	}
}

func TestScript_Schedule(t *testing.T) {
	type fields struct {
		path     string
		schedule string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "with schedule",
			fields: fields{schedule: "@every 5s"},
			want:   "@every 5s",
		},
		{
			name: "without schedule",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Script{
				path:     tt.fields.path,
				schedule: tt.fields.schedule,
			}
			if got := s.Schedule(); got != tt.want {
				t.Errorf("Script.Schedule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScript_Execute(t *testing.T) {
	var validScriptOutput scriptOutput
	err := json.Unmarshal([]byte(jsonOut), &validScriptOutput)
	require.NoError(t, err)
	validSensors := make([]sensor.Details, 0, len(validScriptOutput.Sensors))
	for _, s := range validScriptOutput.Sensors {
		validSensors = append(validSensors, sensor.Details(&s))
	}

	type fields struct {
		path     string
		schedule string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []sensor.Details
		wantErr bool
	}{
		{
			name:   "valid executable",
			fields: fields{path: "/usr/bin/echo" + " " + jsonOut},
			want:   validSensors,
		},
		{
			name:    "invalid executable",
			fields:  fields{path: "/does/not/exist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Script{
				path:     tt.fields.path,
				schedule: tt.fields.schedule,
			}
			got, err := s.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Script.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Script.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
