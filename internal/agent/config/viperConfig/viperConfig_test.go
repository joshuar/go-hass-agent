// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package viperconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestViperConfig_Get(t *testing.T) {
	var key, value string
	key = "testKey"
	value = "testValue"
	v, err := New(t.TempDir())
	assert.Nil(t, err)
	if err := v.Set(key, value); err != nil {
		t.Fail()
	}

	type fields struct {
		store *viper.Viper
		path  string
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "key exists",
			fields:  fields{store: v.store},
			args:    args{key: key, value: &value},
			wantErr: false,
		},
		{
			name:    "key does not exist",
			fields:  fields{store: v.store},
			args:    args{key: "notAKey", value: &value},
			wantErr: true,
		},
		{
			name:    "invalid value",
			fields:  fields{store: v.store},
			args:    args{key: key, value: &struct{}{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ViperConfig{
				store: tt.fields.store,
				path:  tt.fields.path,
			}
			if err := c.Get(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("ViperConfig.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestViperConfig_Set(t *testing.T) {
	var key, value string
	key = "testKey"
	value = "testValue"
	v, err := New(t.TempDir())
	assert.Nil(t, err)

	type fields struct {
		store *viper.Viper
		path  string
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "valid value",
			fields:  fields{store: v.store},
			args:    args{key: key, value: value},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ViperConfig{
				store: tt.fields.store,
				path:  tt.fields.path,
			}
			if err := c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("ViperConfig.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestViperConfig_Delete(t *testing.T) {
	type fields struct {
		store *viper.Viper
		path  string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ViperConfig{
				store: tt.fields.store,
				path:  tt.fields.path,
			}
			if err := c.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("ViperConfig.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestViperConfig_StoragePath(t *testing.T) {
	basePath := filepath.Join(os.Getenv("HOME"), ".config", "go-hass-agent")
	goodPath := filepath.Join(basePath, "goodPath")
	v, err := New(t.TempDir())
	assert.Nil(t, err)

	type fields struct {
		store *viper.Viper
		path  string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "valid path",
			fields:  fields{store: v.store, path: basePath},
			args:    args{id: "goodPath"},
			want:    goodPath,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ViperConfig{
				store: tt.fields.store,
				path:  tt.fields.path,
			}
			got, err := c.StoragePath(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ViperConfig.StoragePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ViperConfig.StoragePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	testPath := t.TempDir()
	v := &ViperConfig{
		store: viper.New(),
		path:  testPath,
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *ViperConfig
		wantErr bool
	}{
		{
			name:    "valid path",
			args:    args{path: testPath},
			want:    v,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("New() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_createDir(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createDir(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("createDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
