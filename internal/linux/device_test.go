// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,dupl
package linux

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/stretchr/testify/require"

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
	case !reflect.DeepEqual(got.distro, want.distro):
		t.Error("distro does not match")
		return false
	case !reflect.DeepEqual(got.distroVersion, want.distroVersion):
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
	withoutOSRelease.distro = unknownDistro
	withoutOSRelease.distroVersion = unknownDistroVersion

	withOSRelease := baseDev
	withOSRelease.distro, withOSRelease.distroVersion = GetDistroID()

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
		want *mqtthass.Device
		name string
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

//revive:disable:unhandled-error
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

func TestGetDistroID(t *testing.T) {
	versionID := "9.9"
	id := "testdistro"
	goodFile := filepath.Join(t.TempDir(), "goodfile")
	fh, err := os.Create(goodFile)
	require.NoError(t, err)
	fmt.Fprintln(fh, `VERSION_ID="`+versionID+`"`)
	fmt.Fprintln(fh, `ID="`+id+`"`)
	fh.Close()

	tests := []struct {
		name          string
		wantID        string
		wantVersionid string
		osReleaseFile string
	}{
		{
			name:          "File exists",
			wantID:        id,
			wantVersionid: versionID,
			osReleaseFile: goodFile,
		},
		{
			name:          "File does not exist.",
			wantID:        unknownDistro,
			wantVersionid: unknownDistroVersion,
			osReleaseFile: "/dev/null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whichdistro.OSReleaseFile = tt.osReleaseFile
			gotID, gotVersionid := GetDistroID()
			if gotID != tt.wantID {
				t.Errorf("GetDistroID() gotId = %v, want %v", gotID, tt.wantID)
			}
			if gotVersionid != tt.wantVersionid {
				t.Errorf("GetDistroID() gotVersionid = %v, want %v", gotVersionid, tt.wantVersionid)
			}
		})
	}
}

func TestGetDistroDetails(t *testing.T) {
	version := "9.9 (note)"
	name := "Test Distro"
	goodFile := filepath.Join(t.TempDir(), "goodfile")
	fh, err := os.Create(goodFile)
	require.NoError(t, err)
	fmt.Fprintln(fh, `VERSION="`+version+`"`)
	fmt.Fprintln(fh, `NAME="`+name+`"`)
	fh.Close()

	tests := []struct {
		name          string
		wantName      string
		wantVersion   string
		osReleaseFile string
	}{
		{
			name:          "File exists",
			wantName:      name,
			wantVersion:   version,
			osReleaseFile: goodFile,
		},
		{
			name:          "File does not exist.",
			wantName:      unknownDistro,
			wantVersion:   unknownDistroVersion,
			osReleaseFile: "/dev/null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whichdistro.OSReleaseFile = tt.osReleaseFile
			gotName, gotVersion := GetDistroDetails()
			if gotName != tt.wantName {
				t.Errorf("GetDistroDetails() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("GetDistroDetails() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}
