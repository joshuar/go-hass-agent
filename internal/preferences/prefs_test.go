// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	_ "embed"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetVersion(tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetDeviceID(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetDeviceName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeviceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestAPIURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetRestAPIURL(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RestAPIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudhookURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetCloudhookURL(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CloudhookURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteUIURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetRemoteUIURL(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoteUIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecret(t *testing.T) {
	type args struct {
		secret string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetSecret(tt.args.secret); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Secret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHost(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetHost(tt.args.host); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Host() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetToken(tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebhookID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetWebhookID(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WebhookID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebsocketURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetWebsocketURL(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WebsocketURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistered(t *testing.T) {
	type args struct {
		status bool
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetRegistered(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMQTTEnabled(t *testing.T) {
	type args struct {
		status bool
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetMQTTEnabled(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MQTTEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMQTTServer(t *testing.T) {
	type args struct {
		server string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetMQTTServer(tt.args.server); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MQTTServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMQTTUser(t *testing.T) {
	type args struct {
		user string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetMQTTUser(tt.args.user); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MQTTUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMQTTPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		args args
		want Preference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetMQTTPassword(tt.args.password); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MQTTPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultPreferences(t *testing.T) {
	tests := []struct {
		name string
		want *Preferences
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
	assert.Nil(t, err)

	tests := []struct {
		name    string
		want    *Preferences
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
			assert.Equal(t, testPrefs.DeviceName, "testDevice")
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
	os.RemoveAll(missingTempPath)
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
