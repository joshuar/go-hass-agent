// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,lll,nlreturn,paralleltest,wsl
//revive:disable:unused-receiver
package scripts

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
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
	type fields struct {
		Output   chan sensor.Details
		path     string
		schedule string
	}
	tests := []struct {
		want    *scriptOutput
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "valid executable",
			fields: fields{path: "/usr/bin/echo" + " " + jsonOut},
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
				Output:   tt.fields.Output,
				path:     tt.fields.path,
				schedule: tt.fields.schedule,
			}
			_, err := s.execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Script.execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Script.execute() = %v, want %v", got, tt.want)
			// }
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

//nolint:containedctx
func TestNewScript(t *testing.T) {
	type args struct {
		ctx  context.Context
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
			args: args{ctx: context.TODO(), path: "/usr/bin/echo" + " " + jsonOut},
			want: &Script{
				path:     "/usr/bin/echo" + " " + jsonOut,
				schedule: "@every 5s",
			},
		},
		{
			name:    "invalid script",
			args:    args{ctx: context.TODO(), path: "/does/not/exist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScript(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.Equal(t, tt.want.path, got.path)
				assert.Equal(t, tt.want.schedule, got.schedule)
			}
		})
	}
}

//nolint:containedctx
func TestFindScripts(t *testing.T) {
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
			args: args{ctx: context.TODO(), path: "internal/scripts/test_files"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindScripts(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}
