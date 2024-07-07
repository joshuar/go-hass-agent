// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,exhaustruct,wsl
package device

import (
	"context"
	"reflect"
	"testing"

	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

func compareDevice(t *testing.T, got, want *Device) bool {
	t.Helper()

	switch {
	case !reflect.DeepEqual(got.appName, want.appName):
		t.Error("appName does not match")

		return false
	case !reflect.DeepEqual(got.appVersion, want.appVersion):
		t.Error("appVersion does not match")

		return false
	case !reflect.DeepEqual(got.osName, want.osName):
		t.Error("distro does not match")

		return false
	case !reflect.DeepEqual(got.osVersion, want.osVersion):
		t.Error("distroVersion does not match")

		return false
	case !reflect.DeepEqual(got.hostname, want.hostname):
		t.Error("hostname does not match")

		return false
	case !reflect.DeepEqual(got.hwModel, want.hwModel):
		t.Error("hwModel does not match")

		return false
	case !reflect.DeepEqual(got.hwVendor, want.hwVendor):
		t.Error("hwVendor does not match")

		return false
	}

	return true
}

func compareMQTTDevice(t *testing.T, got, want *mqtthass.Device) bool {
	t.Helper()

	switch {
	case !reflect.DeepEqual(got.Name, want.Name):
		t.Error("name does not match")

		return false
	case !reflect.DeepEqual(got.URL, want.URL):
		t.Error("URL does not match")

		return false
	case !reflect.DeepEqual(got.SWVersion, want.SWVersion):
		t.Error("SWVersion does not match")

		return false
	case !reflect.DeepEqual(got.Manufacturer, want.Manufacturer):
		t.Error("Manufacturer does not match")

		return false
	case !reflect.DeepEqual(got.Model, want.Model):
		t.Error("Model does not match")

		return false
	}

	return true
}

func TestNewDevice(t *testing.T) {
	origOSRelease := whichdistro.OSReleaseFile
	origAltOSRelease := whichdistro.OSReleaseAltFile

	baseDev := Device{
		appName:    preferences.AppName,
		appVersion: preferences.AppVersion,
		deviceID:   getDeviceID(),
		hostname:   getHostname(),
	}
	baseDev.hwModel, baseDev.hwVendor = getHWProductInfo()

	withoutOSRelease := baseDev
	withoutOSRelease.osName = unknownDistro
	withoutOSRelease.osVersion = unknownDistroVersion

	withOSRelease := baseDev
	withOSRelease.osName, withOSRelease.osVersion = getOSID()

	type args struct {
		name           string
		version        string
		osReleaseFiles []string
	}

	tests := []struct {
		want *Device
		name string
		args args
	}{
		{
			name: "with OS Release",
			args: args{
				name:           preferences.AppName,
				version:        preferences.AppVersion,
				osReleaseFiles: []string{whichdistro.OSReleaseFile, whichdistro.OSReleaseAltFile},
			},
			want: &withOSRelease,
		},
		{
			name: "without OS Release",
			args: args{
				name:           preferences.AppName,
				version:        preferences.AppVersion,
				osReleaseFiles: []string{"", ""},
			},
			want: &withoutOSRelease,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whichdistro.OSReleaseFile = tt.args.osReleaseFiles[0]
			whichdistro.OSReleaseAltFile = tt.args.osReleaseFiles[1]

			if got := New(tt.args.name, tt.args.version); !compareDevice(t, got, tt.want) {
				t.Errorf("NewDevice() = %v, want %v", got, tt.want)
			}

			whichdistro.OSReleaseFile = origOSRelease
			whichdistro.OSReleaseAltFile = origAltOSRelease
		})
	}
}

//nolint:containedctx
func TestMQTTDevice(t *testing.T) {
	dev := New(preferences.AppName, preferences.AppVersion)
	mqttDev := &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
	ctx := preferences.ContextSetPrefs(context.TODO(), preferences.DefaultPreferences())

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		args args
		want *mqtthass.Device
		name string
	}{
		{
			name: "default",
			want: mqttDev,
			args: args{ctx: ctx},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MQTTDeviceInfo(tt.args.ctx); !compareMQTTDevice(t, got, tt.want) {
				t.Errorf("MQTTDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}
