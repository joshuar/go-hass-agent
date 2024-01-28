// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package fyneconfig

import (
	"os"
	"reflect"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func TestNewFyneConfig(t *testing.T) {
	app := app.NewWithID("org.github.joshuar.go-hass-agent-test")
	rootPath := app.Storage().RootURI()
	defer os.RemoveAll(rootPath.Path())

	tests := []struct {
		name string
		want *FyneConfig
	}{
		{
			name: "default test",
			want: &FyneConfig{
				prefs: app.Preferences(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFyneConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFyneConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFyneConfig_Get(t *testing.T) {
	app := app.NewWithID("org.github.joshuar.go-hass-agent-test")
	rootPath := app.Storage().RootURI()
	defer os.RemoveAll(rootPath.Path())
	app.Preferences().SetString("aString", "aValue")

	var stringValue, missingValue string
	var boolValue bool

	type fields struct {
		prefs fyne.Preferences
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
			name:    "get string",
			fields:  fields{prefs: app.Preferences()},
			args:    args{key: "aString", value: &stringValue},
			wantErr: false,
		},
		{
			name:    "get bool",
			fields:  fields{prefs: app.Preferences()},
			args:    args{key: "aBool", value: &boolValue},
			wantErr: false,
		},
		{
			name:    "missing string",
			fields:  fields{prefs: app.Preferences()},
			args:    args{key: "aMissingKey", value: &missingValue},
			wantErr: true,
		},
		{
			name:    "unsupported value",
			fields:  fields{prefs: app.Preferences()},
			args:    args{key: "aSlice", value: make([]string, 3)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &FyneConfig{
				prefs: tt.fields.prefs,
			}
			if err := c.Get(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("FyneConfig.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFyneConfig_Set(t *testing.T) {
	app := app.NewWithID("org.github.joshuar.go-hass-agent-test")
	rootPath := app.Storage().RootURI()
	defer os.RemoveAll(rootPath.Path())

	type fields struct {
		prefs fyne.Preferences
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
			name:   "set string",
			fields: fields{prefs: app.Preferences()},
			args:   args{key: "aString", value: "aString"},
		},
		{
			name:   "set bool",
			fields: fields{prefs: app.Preferences()},
			args:   args{key: "aBool", value: true},
		},
		{
			name:    "set unsuppported",
			fields:  fields{prefs: app.Preferences()},
			args:    args{key: "aSlice", value: make([]string, 3)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &FyneConfig{
				prefs: tt.fields.prefs,
			}
			if err := c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("FyneConfig.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFyneConfig_Delete(t *testing.T) {
	type fields struct {
		prefs fyne.Preferences
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
			c := &FyneConfig{
				prefs: tt.fields.prefs,
			}
			if err := c.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("FyneConfig.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFyneConfig_StoragePath(t *testing.T) {
	app := app.NewWithID("org.github.joshuar.go-hass-agent-test")
	rootPath := app.Storage().RootURI()
	defer os.RemoveAll(rootPath.Path())

	type fields struct {
		prefs fyne.Preferences
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
			name:   "default test",
			fields: fields{prefs: app.Preferences()},
			args:   args{id: "storageID"},
			want:   rootPath.Path() + "/storageID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &FyneConfig{
				prefs: tt.fields.prefs,
			}
			got, err := c.StoragePath(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FyneConfig.StoragePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FyneConfig.StoragePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
