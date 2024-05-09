// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"os"
	"reflect"
	"testing"

	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

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
	withoutOSRelease.distro = "Unknown Distro"
	withoutOSRelease.distroVersion = "Unknown Version"

	osReleaseInfo, err := whichdistro.GetOSRelease()
	assert.Nil(t, err)
	withOSRelease := baseDev
	withOSRelease.distro = osReleaseInfo["ID"]
	withOSRelease.distroVersion = osReleaseInfo["VERSION_ID"]

	type args struct {
		name           string
		version        string
		osReleaseFiles []string
	}
	tests := []struct {
		name string
		args args
		want *Device
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
			if got := NewDevice(tt.args.name, tt.args.version); !compareDevice(t, got, tt.want) {
				t.Errorf("NewDevice() = %v, want %v", got, tt.want)
			}
			whichdistro.OSReleaseFile = origOSRelease
			whichdistro.OSReleaseAltFile = origAltOSRelease
		})
	}
}

func TestMQTTDevice(t *testing.T) {
	dev := NewDevice(preferences.AppName, preferences.AppVersion)
	mqttDev := &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
	tests := []struct {
		name string
		want *mqtthass.Device
	}{
		{
			name: "default",
			want: mqttDev,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MQTTDevice(); !compareMQTTDevice(t, got, tt.want) {
				t.Errorf("MQTTDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindPortal(t *testing.T) {
	type args struct {
		setup func()
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "KDE",
			args: args{
				setup: func() { os.Setenv("XDG_CURRENT_DESKTOP", "KDE") },
			},
			want: "org.freedesktop.impl.portal.desktop.kde",
		},
		{
			name: "GNOME",
			args: args{
				setup: func() { os.Setenv("XDG_CURRENT_DESKTOP", "GNOME") },
			},
			want: "org.freedesktop.impl.portal.desktop.gtk",
		},
		{
			name: "Unknown",
			args: args{
				setup: func() { os.Setenv("XDG_CURRENT_DESKTOP", "UNKNOWN") },
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			if got := FindPortal(); got != tt.want {
				t.Errorf("FindPortal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareDevice(t *testing.T, a, b *Device) bool {
	switch {
	case !reflect.DeepEqual(a.appName, b.appName):
		t.Error("appName does not match")
		return false
	case !reflect.DeepEqual(a.appVersion, b.appVersion):
		t.Error("appVersion does not match")
		return false
	case !reflect.DeepEqual(a.distro, b.distro):
		t.Error("distro does not match")
		return false
	case !reflect.DeepEqual(a.distroVersion, b.distroVersion):
		t.Error("distroVersion does not match")
		return false
	case !reflect.DeepEqual(a.hostname, b.hostname):
		t.Error("hostname does not match")
		return false
	case !reflect.DeepEqual(a.hwModel, b.hwModel):
		t.Error("hwModel does not match")
		return false
	case !reflect.DeepEqual(a.hwVendor, b.hwVendor):
		t.Error("hwVendor does not match")
		return false
	}
	return true
}

func compareMQTTDevice(t *testing.T, a, b *mqtthass.Device) bool {
	switch {
	case !reflect.DeepEqual(a.Name, b.Name):
		t.Error("name does not match")
		return false
	case !reflect.DeepEqual(a.URL, b.URL):
		t.Error("URL does not match")
		return false
	case !reflect.DeepEqual(a.SWVersion, b.SWVersion):
		t.Error("SWVersion does not match")
		return false
	case !reflect.DeepEqual(a.Manufacturer, b.Manufacturer):
		t.Error("Manufacturer does not match")
		return false
	case !reflect.DeepEqual(a.Model, b.Model):
		t.Error("Model does not match")
		return false
	}
	return true
}
