// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:dupl,exhaustruct,paralleltest
package registry

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

var testSensorMap = map[string]metadata{
	"registeredSensor":   {Disabled: false, Registered: true},
	"unRegisteredSensor": {Disabled: false, Registered: false},
	"disabledSensor":     {Disabled: true, Registered: true},
}

func newTestRegistry(t *testing.T) *gobRegistry {
	t.Helper()

	registryPath = t.TempDir()

	testRegistry, err := Load()
	require.NoError(t, err)

	testRegistry.sensors = testSensorMap

	err = testRegistry.write()
	require.NoError(t, err)

	return testRegistry
}

func Test_gobRegistry_write(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
	}

	type args struct {
		path string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "default",
			args:    args{path: t.TempDir()},
			fields:  fields{sensors: testSensorMap},
			wantErr: false,
		},
		{
			name:    "invalid path",
			args:    args{path: "/nonexistent"},
			fields:  fields{sensors: testSensorMap},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &gobRegistry{
				sensors: tt.fields.sensors,
			}
			registryPath = tt.args.path

			if err := registry.write(); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gobRegistry_read(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
	}

	type args struct {
		path string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "default",
			args:    args{path: t.TempDir()},
			fields:  fields{sensors: testSensorMap},
			wantErr: false,
		},
		{
			name:    "invalid path",
			args:    args{path: "/nonexistent"},
			fields:  fields{sensors: testSensorMap},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &gobRegistry{
				sensors: tt.fields.sensors,
			}
			registryPath = tt.args.path

			if err := registry.read(); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.read() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gobRegistry_IsDisabled(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
	}

	type args struct {
		id string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "disabled sensor",
			fields: fields{sensors: testSensorMap},
			args:   args{id: "disabledSensor"},
			want:   true,
		},
		{
			name:   "not disabled sensor",
			fields: fields{sensors: testSensorMap},
			args:   args{id: "registeredSensor"},
			want:   false,
		},
		{
			name:   "not found",
			fields: fields{sensors: testSensorMap},
			args:   args{id: "nonexistent"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newTestRegistry(t)

			if got := registry.IsDisabled(tt.args.id); got != tt.want {
				t.Errorf("gobRegistry.IsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gobRegistry_IsRegistered(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
	}

	type args struct {
		id string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "registered sensor",
			fields: fields{sensors: testSensorMap},
			args:   args{id: "registeredSensor"},
			want:   true,
		},
		{
			name:   "not registered sensor",
			fields: fields{sensors: testSensorMap},
			args:   args{id: "unRegistered"},
			want:   false,
		},
		{
			name:   "not found",
			fields: fields{sensors: testSensorMap},
			args:   args{id: "nonexistent"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestRegistry(t)
			if got := g.IsRegistered(tt.args.id); got != tt.want {
				t.Errorf("gobRegistry.IsRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gobRegistry_SetDisabled(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
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
			fields:  fields{sensors: testSensorMap},
			args:    args{id: "disabledSensor", value: false},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newTestRegistry(t)

			if err := registry.SetDisabled(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.SetDisabled() error = %v, wantErr %v", err, tt.wantErr)
			}

			if registry.IsDisabled(tt.args.id) != tt.args.value {
				t.Error("gobRegistry.SetDisabled() not changed")
			}
		})
	}
}

func Test_gobRegistry_SetRegistered(t *testing.T) {
	type fields struct {
		sensors map[string]metadata
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
			fields:  fields{sensors: testSensorMap},
			args:    args{id: "unRegisteredSensor", value: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newTestRegistry(t)

			if err := registry.SetRegistered(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("gobRegistry.SetRegistered() error = %v, wantErr %v", err, tt.wantErr)
			}

			if registry.IsRegistered(tt.args.id) != tt.args.value {
				t.Error("gobRegistry.SetRegistered() not changed")
			}
		})
	}
}

func TestLoad(t *testing.T) {
	type args struct {
		path string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "good path",
			args:    args{path: t.TempDir()},
			wantErr: false,
		},
		{
			name:    "bad path",
			args:    args{path: "/nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registryPath = tt.args.path

			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			registryPath = filepath.Join(xdg.ConfigHome, "sensorRegistry")
		})
	}
}
