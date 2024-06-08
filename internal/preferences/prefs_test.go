// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest
package preferences

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func Test_defaultPreferences(t *testing.T) {
	tests := []struct {
		want *Preferences
		name string
	}{
		{
			name: "default returned",
			want: &Preferences{
				Version: AppVersion,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultPreferences()
			assert.Equal(t, got.Version, AppVersion)
		})
	}
}

func TestLoad(t *testing.T) {
	missingPrefsDir := t.TempDir()

	testServer := "http://test.host:9999"
	existingPrefsDir := t.TempDir()
	existingPrefs := &Preferences{
		Host:         testServer,
		Token:        "testToken",
		WebhookID:    "testID",
		RestAPIURL:   testServer,
		WebsocketURL: testServer,
		DeviceID:     "testID",
		DeviceName:   "testDevice",
		Version:      "6.4.0",
		Registered:   true,
	}
	err := write(existingPrefs, filepath.Join(existingPrefsDir, preferencesFile))
	require.NoError(t, err)

	tests := []struct {
		want    *Preferences
		name    string
		wantErr bool
	}{
		{
			name:    "missing",
			want:    defaultPreferences(),
			wantErr: true,
		},
		{
			name:    "existing",
			want:    existingPrefs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "missing":
				SetPath(missingPrefsDir)
			case "existing":
				SetPath(existingPrefsDir)
			}

			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			assert.Equal(t, got.DeviceName, tt.want.DeviceName)
		})
	}
}

func TestSave(t *testing.T) {
	SetPath(t.TempDir())

	testServer := "http://test.host:9999"

	requiredPrefs := []Preference{
		SetHost(testServer),
		SetToken("testToken"),
		SetCloudhookURL(""),
		SetRemoteUIURL(""),
		SetWebhookID("testID"),
		SetSecret(""),
		SetRestAPIURL(testServer),
		SetWebsocketURL(testServer),
		SetDeviceName("testDevice"),
		SetDeviceID("testID"),
		SetVersion("6.4.0"),
		SetRegistered(true),
	}

	missingPrefs := []Preference{
		SetHost(testServer),
		SetToken("testToken"),
	}

	type args struct {
		setters []Preference
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "save defaults (and fail)",
			wantErr: true,
		},
		{
			name:    "save some (and fail)",
			args:    args{setters: missingPrefs},
			wantErr: true,
		},
		{
			name:    "save all",
			args:    args{setters: requiredPrefs},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Save(tt.args.setters...); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_set(t *testing.T) {
	testPrefs := defaultPreferences()
	testSetter := func(p *Preferences) error {
		p.DeviceName = "testDevice"

		return nil
	}

	type args struct {
		prefs   *Preferences
		setters []Preference
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "single setter",
			args: args{prefs: testPrefs, setters: []Preference{testSetter}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := set(tt.args.prefs, tt.args.setters...); (err != nil) != tt.wantErr {
				t.Errorf("set() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, "testDevice", testPrefs.DeviceName)
		})
	}
}

func Test_write(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "testpreferences.toml")

	type args struct {
		prefs *Preferences
		file  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{prefs: defaultPreferences(), file: testFile},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := write(tt.args.prefs, tt.args.file); (err != nil) != tt.wantErr {
				t.Errorf("write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkPath(t *testing.T) {
	missingTempPath := t.TempDir()

	err := os.RemoveAll(missingTempPath)
	require.NoError(t, err)

	existingTempPath := t.TempDir()

	type args struct {
		path string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "missing",
			args:    args{path: missingTempPath},
			wantErr: false,
		},
		{
			name:    "existing",
			args:    args{path: existingTempPath},
			wantErr: false,
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
