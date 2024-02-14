// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package registry

import (
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_jsonFilesRegistry_get(t *testing.T) {
	mockMap := sync.Map{}
	mockMap.Store("disabled", metadata{Disabled: true, Registered: false})
	mockMap.Store("registered", metadata{Disabled: false, Registered: true})

	type fields struct {
		sensors sync.Map
		path    string
	}
	type args struct {
		id        string
		valueType state
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "disabled sensor",
			fields: fields{sensors: mockMap},
			args:   args{id: "disabled", valueType: disabledState},
			want:   true,
		},
		{
			name:   "registered sensor",
			fields: fields{sensors: mockMap},
			args:   args{id: "registered", valueType: registeredState},
			want:   true,
		},
		{
			name:   "unknown sensor",
			fields: fields{sensors: mockMap},
			args:   args{id: "unknown", valueType: registeredState},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if got := j.get(tt.args.id, tt.args.valueType); got != tt.want {
				t.Errorf("jsonFilesRegistry.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonFilesRegistry_IsDisabled(t *testing.T) {
	mockMap := sync.Map{}
	mockMap.Store("disabled", metadata{Disabled: true, Registered: false})

	type fields struct {
		sensors sync.Map
		path    string
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
			name:   "successful test",
			fields: fields{sensors: mockMap},
			args:   args{id: "disabled"},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if got := j.IsDisabled(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jsonFilesRegistry.IsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonFilesRegistry_IsRegistered(t *testing.T) {
	mockMap := sync.Map{}
	mockMap.Store("registered", metadata{Disabled: false, Registered: true})

	type fields struct {
		sensors sync.Map
		path    string
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
			name:   "successful test",
			fields: fields{sensors: mockMap},
			args:   args{id: "registered"},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if got := j.IsRegistered(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jsonFilesRegistry.IsRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonFilesRegistry_set(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-hass-agent-test-*")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	mockMap := sync.Map{}
	mockMap.Store("existing", metadata{Disabled: false, Registered: false})

	type fields struct {
		sensors sync.Map
		path    string
	}
	type args struct {
		id        string
		valueType state
		value     bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "new sensor",
			fields:  fields{sensors: mockMap, path: tmpDir},
			args:    args{id: "new", valueType: registeredState, value: true},
			wantErr: false,
		},
		{
			name:    "existing sensor",
			fields:  fields{sensors: mockMap, path: tmpDir},
			args:    args{id: "existing", valueType: disabledState, value: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if err := j.set(tt.args.id, tt.args.valueType, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("jsonFilesRegistry.set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_jsonFilesRegistry_write(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-hass-agent-test-*")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	mockMap := sync.Map{}
	mockMap.Store("existing", metadata{Disabled: false, Registered: false})

	type fields struct {
		sensors sync.Map
		path    string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "existing sensor",
			fields:  fields{sensors: mockMap, path: tmpDir},
			args:    args{id: "existing"},
			wantErr: false,
		},
		{
			name:    "nonexisting sensor",
			fields:  fields{sensors: mockMap, path: tmpDir},
			args:    args{id: "nonexisting"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if err := j.write(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("jsonFilesRegistry.write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_jsonFilesRegistry_SetDisabled(t *testing.T) {
	type fields struct {
		sensors sync.Map
		path    string
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
		// * same functionality tested by get
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if err := j.SetDisabled(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("jsonFilesRegistry.SetDisabled() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_jsonFilesRegistry_SetRegistered(t *testing.T) {
	type fields struct {
		sensors sync.Map
		path    string
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
		// * same functionality tested by get
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &jsonFilesRegistry{
				sensors: tt.fields.sensors,
				path:    tt.fields.path,
			}
			if err := j.SetRegistered(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("jsonFilesRegistry.SetRegistered() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewJsonFilesRegistry(t *testing.T) {
	SetPath(t.TempDir())

	tests := []struct {
		name    string
		want    *jsonFilesRegistry
		wantErr bool
	}{
		{
			name:    "default",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJsonFilesRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// * don't really care about equivalence for test
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewJsonFilesRegistry() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_parseFile(t *testing.T) {
	goodPath, err := os.MkdirTemp("", "go-hass-agent-Test_parseFile_good-*")
	assert.Nil(t, err)
	defer os.RemoveAll(goodPath)

	goodFilePath := goodPath + "/good.json"
	err = os.WriteFile(goodFilePath, []byte(`{"Registered":true,"Disabled":false}`), 0o644)
	assert.Nil(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 metadata
	}{
		{
			name:  "successful parse",
			args:  args{path: goodFilePath},
			want:  "good",
			want1: metadata{Registered: true, Disabled: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseFile(tt.args.path)
			if got != tt.want {
				t.Errorf("parseFile() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseFile() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
