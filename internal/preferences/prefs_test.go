// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl,nlreturn,dupl,varnamelen
//revive:disable:unused-receiver,comment-spacings
package preferences

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

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
	case !reflect.DeepEqual(got.file, want.file):
		t.Error("file does not match")
		return false
	}
	return true
}

func TestPreferences_Validate(t *testing.T) {
	validPrefs := DefaultPreferences(filepath.Join(t.TempDir(), preferencesFile))

	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		file         string
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
				file:         tt.fields.file,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPreferences_Save(t *testing.T) {
	validPrefs := DefaultPreferences(filepath.Join(t.TempDir(), preferencesFile))

	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		file         string
		Registered   bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid preferences",
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
				file:         validPrefs.file,
			},
		},
		{
			name:    "invalid preferences",
			wantErr: true,
		},
		{
			name: "unwriteable preferences path",
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
				Hass:         validPrefs.Hass,
				Device:       validPrefs.Device,
				Version:      AppVersion,
				file:         "/",
			},
			wantErr: true,
		},
		{
			name: "missing preferences path",
			fields: fields{
				MQTT:         validPrefs.MQTT,
				Registration: validPrefs.Registration,
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
				file:         tt.fields.file,
			}
			if err := p.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPreferences_GetMQTTPreferences(t *testing.T) {
	validPrefs := DefaultPreferences(filepath.Join(t.TempDir(), preferencesFile))

	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		file         string
		Registered   bool
	}
	tests := []struct {
		want   *MQTT
		name   string
		fields fields
	}{
		{
			name: "valid MQTT prefs",
			fields: fields{
				MQTT: validPrefs.MQTT,
			},
			want: validPrefs.MQTT,
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
				file:         tt.fields.file,
			}
			if got := p.GetMQTTPreferences(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Preferences.GetMQTTPreferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	noFile := filepath.Join(t.TempDir(), preferencesFile)
	invalidFile := filepath.Join(t.TempDir(), preferencesFile)
	err := os.WriteFile(invalidFile, []byte(`invalid`), 0o600)
	require.NoError(t, err)
	existingPrefs := DefaultPreferences(filepath.Join(t.TempDir(), preferencesFile))
	err = existingPrefs.Save()
	require.NoError(t, err)

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
			args:        args{path: filepath.Dir(noFile)},
			want:        DefaultPreferences(noFile),
			wantErr:     true,
			wantErrType: ErrNoPreferences,
		},
		{
			name:        "invalid file",
			args:        args{path: filepath.Dir(invalidFile)},
			want:        DefaultPreferences(invalidFile),
			wantErr:     true,
			wantErrType: ErrFileContents,
		},
		{
			name: "existing file",
			args: args{path: filepath.Dir(existingPrefs.file)},
			want: existingPrefs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.ErrorIs(t, err, tt.wantErrType)
			if !preferencesEqual(t, got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset(t *testing.T) {
	existingPrefs := DefaultPreferences(filepath.Join(t.TempDir(), preferencesFile))
	err := existingPrefs.Save()
	require.NoError(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid path",
			args: args{path: filepath.Dir(existingPrefs.file)},
		},
		{
			name:    "invalid path",
			args:    args{path: filepath.Join(t.TempDir(), "nonexistent")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Reset(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
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
