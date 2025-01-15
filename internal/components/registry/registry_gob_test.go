// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package registry

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockSensors = map[string]metadata{
	"disabledSensor":   {Disabled: true, Registered: true},
	"registeredSensor": {Disabled: false, Registered: true},
}

func newMockReg(t *testing.T, path string) *gobRegistry {
	t.Helper()

	mockReg, err := Load(path)
	require.NoError(t, err)
	mockReg.sensors = mockSensors
	err = mockReg.write()
	require.NoError(t, err)
	return mockReg
}

func Test_gobRegistry_write(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
		file    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid path",
			fields: fields{sensors: mockSensors, file: filepath.Join(t.TempDir(), registryFile)},
		},
		{
			name:    "invalid path",
			fields:  fields{sensors: mockSensors, file: filepath.Join(t.TempDir(), "nonexistent", registryFile)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gobRegistry{
				sensors: tt.fields.sensors,
				file:    tt.fields.file,
			}
			if err := g.write(); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.write() error c= %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gobRegistry_read(t *testing.T) {
	mockReg := newMockReg(t, t.TempDir())

	invalidRegistry := filepath.Join(t.TempDir(), registryFile)
	err := os.WriteFile(invalidRegistry, []byte(`invalid`), 0o600)
	require.NoError(t, err)

	type fields struct {
		sensors map[string]metadata
		file    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid file",
			fields: fields{file: mockReg.file},
		},
		{
			name:    "invalid file",
			fields:  fields{file: filepath.Join(t.TempDir(), "nonexistent", registryFile)},
			wantErr: true,
		},
		{
			name:    "invalid contents",
			fields:  fields{file: invalidRegistry},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gobRegistry{
				sensors: tt.fields.sensors,
				file:    tt.fields.file,
			}
			if err := g.read(); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.read() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, mockReg.sensors, g.sensors)
			}
		})
	}
}

func Test_gobRegistry_IsDisabled(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
		file    string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   bool
	}{
		{
			name:   "disabled sensor",
			fields: fields{sensors: mockSensors},
			args:   args{id: "disabledSensor"},
			want:   true,
		},
		{
			name:   "not disabled sensor",
			fields: fields{sensors: mockSensors},
			args:   args{id: "registeredSensor"},
			want:   false,
		},
		{
			name:   "not found",
			fields: fields{sensors: mockSensors},
			args:   args{id: "nonexistent"},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gobRegistry{
				sensors: tt.fields.sensors,
				file:    tt.fields.file,
			}
			if got := g.IsDisabled(tt.args.id); got != tt.want {
				t.Errorf("gobRegistry.IsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gobRegistry_IsRegistered(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
		file    string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   bool
	}{
		{
			name:   "registered sensor",
			fields: fields{sensors: mockSensors},
			args:   args{id: "registeredSensor"},
			want:   true,
		},
		{
			name:   "not registered sensor",
			fields: fields{sensors: mockSensors},
			args:   args{id: "unRegistered"},
			want:   false,
		},
		{
			name:   "not found",
			fields: fields{sensors: mockSensors},
			args:   args{id: "nonexistent"},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gobRegistry{
				sensors: tt.fields.sensors,
				file:    tt.fields.file,
			}
			if got := g.IsRegistered(tt.args.id); got != tt.want {
				t.Errorf("gobRegistry.IsRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gobRegistry_SetDisabled(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
		file    string
	}
	type args struct {
		id    string
		value bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "change disabled state",
			fields:  fields{sensors: mockSensors, file: filepath.Join(t.TempDir(), registryFile)},
			args:    args{id: "disabledSensor", value: false},
			wantErr: false,
		},
		{
			name:    "invalid path",
			fields:  fields{sensors: mockSensors, file: filepath.Join(t.TempDir(), "nonexistent", registryFile)},
			args:    args{id: "disabledSensor", value: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gobRegistry{
				sensors: tt.fields.sensors,
				file:    tt.fields.file,
			}
			if err := g.SetDisabled(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.SetDisabled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gobRegistry_SetRegistered(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
		file    string
	}
	type args struct {
		id    string
		value bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "change registered state",
			fields:  fields{sensors: mockSensors, file: filepath.Join(t.TempDir(), registryFile)},
			args:    args{id: "unRegisteredSensor", value: true},
			wantErr: false,
		},
		{
			name:    "invalid path",
			fields:  fields{sensors: mockSensors, file: filepath.Join(t.TempDir(), "nonexistent", registryFile)},
			args:    args{id: "disabledSensor", value: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gobRegistry{
				sensors: tt.fields.sensors,
				file:    tt.fields.file,
			}
			if err := g.SetRegistered(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.SetRegistered() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	goodPath := t.TempDir()
	badPath := "/nonexistent"

	type args struct {
		path string
	}
	tests := []struct {
		want    *gobRegistry
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good path",
			args: args{path: goodPath},
			want: &gobRegistry{
				sensors: make(map[string]metadata),
				file: filepath.Join(
					goodPath,
					"sensorRegistry",
					registryFile),
			},
			wantErr: false,
		},
		{
			name:    "bad path",
			args:    args{path: badPath},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}
