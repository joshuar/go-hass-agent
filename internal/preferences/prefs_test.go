// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl,nlreturn,varnamelen
//revive:disable:function-length
package preferences

import (
	_ "embed"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deviceEqual(t *testing.T, got, want *Device) bool {
	t.Helper()
	switch {
	case !reflect.DeepEqual(got.AppName, want.AppName):
		t.Error("appName does not match")
		return false
	case !reflect.DeepEqual(got.AppVersion, want.AppVersion):
		t.Error("appVersion does not match")
		return false
	case !reflect.DeepEqual(got.OsName, want.OsName):
		t.Error("distro does not match")
		return false
	case !reflect.DeepEqual(got.OsVersion, want.OsVersion):
		t.Errorf("distroVersion does not match: got %s want %s", got.OsVersion, want.OsVersion)
		return false
	case !reflect.DeepEqual(got.Name, want.Name):
		t.Error("hostname does not match")
		return false
	case !reflect.DeepEqual(got.Model, want.Model):
		t.Error("hwModel does not match")
		return false
	case !reflect.DeepEqual(got.Manufacturer, want.Manufacturer):
		t.Error("hwVendor does not match")
		return false
	}
	return true
}

func preferencesEqual(t *testing.T, got, want *Preferences) bool {
	t.Helper()
	switch {
	case !deviceEqual(t, got.Device, want.Device):
		t.Error("device does not match")
		return false
	case !reflect.DeepEqual(got.Hass, want.Hass):
		t.Error("hass preferences do not match")
		return false
	case !reflect.DeepEqual(got.Registration, want.Registration):
		t.Error("registration preferences do not match")
		return false
	case !reflect.DeepEqual(got.MQTT, want.MQTT):
		t.Errorf("mqtt preferences do not match")
		return false
	case !reflect.DeepEqual(got.Version, want.Version):
		t.Error("version does not match")
		return false
	case !reflect.DeepEqual(got.Registered, want.Registered):
		t.Error("registered does not match")
		return false
	}
	return true
}

func TestSetPath(t *testing.T) {
	testPath := t.TempDir()

	type args struct {
		path string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "set a path",
			args: args{path: testPath},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPath(tt.args.path)
			assert.Equal(t, Path(), testPath)
		})
	}
}

func TestSetFile(t *testing.T) {
	testName := "testfile"

	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "set a file",
			args: args{name: testName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetFile(tt.args.name)
			assert.Equal(t, File(), testName)
		})
	}
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "default path",
			want: preferencesPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Path(); got != tt.want {
				t.Errorf("GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "default file",
			want: preferencesFile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := File(); got != tt.want {
				t.Errorf("GetFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// DefaultPeferences does not make sense to test.
// func Test_defaultPreferences(t *testing.T) {}

func TestLoad(t *testing.T) {
	origPath := Path()

	newPreferencesDir := t.TempDir()

	invalidPreferencesDir := t.TempDir()
	err := os.WriteFile(filepath.Join(invalidPreferencesDir, preferencesFile), []byte(`invalid toml`), 0o600)
	require.NoError(t, err)

	existingFileDir := t.TempDir()
	existingPrefs := DefaultPreferences()
	SetPath(existingFileDir)
	err = existingPrefs.Save()
	require.NoError(t, err)
	SetPath(origPath)

	type args struct {
		path string
	}
	tests := []struct {
		wantErrType error
		want        *Preferences
		name        string
		args        args
		wantErr     bool
	}{
		{
			name:        "new file",
			args:        args{path: newPreferencesDir},
			want:        DefaultPreferences(),
			wantErr:     true,
			wantErrType: ErrNoPreferences,
		},
		{
			name:        "invalid file",
			args:        args{path: invalidPreferencesDir},
			want:        DefaultPreferences(),
			wantErr:     true,
			wantErrType: ErrFileContents,
		},
		{
			name: "existing file",
			args: args{path: existingFileDir},
			want: DefaultPreferences(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPath(tt.args.path)
			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.ErrorIs(t, err, tt.wantErrType)
			if !preferencesEqual(t, got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
			SetPath(origPath)
		})
	}
}

func TestPreferences_Validate(t *testing.T) {
	validPrefs := DefaultPreferences()

	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		Registered   bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
			},
		},
		{
			name: "required field missing",
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
			},
			wantErr: true,
		},
		{
			name: "invalid field value",
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: &Registration{Server: "notaurl", Token: "somestring"},
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Preferences{
				mu:           tt.fields.mu,
				MQTT:         tt.fields.MQTT,
				Registration: tt.fields.Registration,
				Hass:         tt.fields.Hass,
				Device:       tt.fields.Device,
				Version:      tt.fields.Version,
				Registered:   tt.fields.Registered,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPreferences_Save(t *testing.T) {
	origPath := Path()
	validPrefs := DefaultPreferences()

	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		Registered   bool
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "valid preferences",
			args: args{path: t.TempDir()},
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
			},
		},
		{
			name:    "invalid preferences",
			args:    args{path: t.TempDir()},
			wantErr: true,
		},
		{
			name: "unwriteable preferences path",
			args: args{path: "/"},
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
			},
			wantErr: true,
		},
		{
			name: "missing preferences path",
			args: args{path: filepath.Join(t.TempDir(), "missing")},
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPath(tt.args.path)
			p := &Preferences{
				mu:           tt.fields.mu,
				MQTT:         tt.fields.MQTT,
				Registration: tt.fields.Registration,
				Hass:         tt.fields.Hass,
				Device:       tt.fields.Device,
				Version:      tt.fields.Version,
				Registered:   tt.fields.Registered,
			}
			if err := p.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			SetPath(origPath)
		})
	}
}

func TestReset(t *testing.T) {
	origPath := Path()
	existsPath := t.TempDir()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exists",
			args: args{path: existsPath},
		},
		{
			name: "does not exist",
			args: args{path: "/doesnotexist"},
		},
		{
			name:    "unwriteable",
			args:    args{path: "/proc/self"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		SetPath(tt.args.path)
		t.Run(tt.name, func(t *testing.T) {
			if err := Reset(); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		assert.NoDirExists(t, tt.args.path)
		SetPath(origPath)
	}
}

func TestMQTT_TopicPrefix(t *testing.T) {
	type fields struct {
		MQTTServer      string
		MQTTUser        string
		MQTTPassword    string
		MQTTTopicPrefix string
		MQTTEnabled     bool
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "no topic set",
			want: MQTTTopicPrefix,
		},
		{
			name:   "custom topic set",
			fields: fields{MQTTTopicPrefix: "testtopic"},
			want:   "testtopic",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MQTT{
				MQTTServer:      tt.fields.MQTTServer,
				MQTTUser:        tt.fields.MQTTUser,
				MQTTPassword:    tt.fields.MQTTPassword,
				MQTTTopicPrefix: tt.fields.MQTTTopicPrefix,
				MQTTEnabled:     tt.fields.MQTTEnabled,
			}
			if got := p.TopicPrefix(); got != tt.want {
				t.Errorf("MQTT.TopicPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exists",
			args: args{path: t.TempDir()},
		},
		{
			name: "does not exist",
			args: args{path: filepath.Join(t.TempDir(), "notexists")},
		},
		{
			name:    "unwriteable",
			args:    args{path: "/notexists"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkPath(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("checkPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
