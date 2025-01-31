// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package whichdistro

import (
	"embed"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/*
var content embed.FS

func resetFiles() {
	OSReleaseFile = "/etc/os-release"
	OSReleaseAltFile = "/usr/lib/os-release"
}

func TestGetOSRelease(t *testing.T) {
	tests := []struct {
		want             OSRelease
		name             string
		osReleaseFile    string
		osAltReleaseFile string
		wantErr          bool
	}{
		{
			name:          "successful",
			osReleaseFile: "./testdata/os-release-fedora",
			wantErr:       false,
		},
		{
			name:             "unsuccessful",
			osReleaseFile:    "/does/not/exist",
			osAltReleaseFile: "/also/does/not/exist",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		OSReleaseFile = tt.osReleaseFile
		OSReleaseAltFile = tt.osAltReleaseFile
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetOSRelease()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOSRelease() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
		resetFiles()
	}
}

func Test_readOSRelease(t *testing.T) {
	var (
		fedora, ubuntu, tumbleweed []byte
		err                        error
	)

	fedora, err = content.ReadFile("testdata/os-release-fedora")
	require.NoError(t, err)

	ubuntu, err = content.ReadFile("testdata/os-release-ubuntu")
	require.NoError(t, err)

	tumbleweed, err = content.ReadFile("testdata/os-release-opensuse-tumbleweed")
	require.NoError(t, err)

	tests := []struct {
		name             string
		osReleaseFile    string
		osAltReleaseFile string
		want             []byte
		wantErr          bool
	}{
		{
			name:          "fedora",
			want:          fedora,
			wantErr:       false,
			osReleaseFile: "testdata/os-release-fedora",
		},
		{
			name:          "ubuntu",
			want:          ubuntu,
			wantErr:       false,
			osReleaseFile: "testdata/os-release-ubuntu",
		},
		{
			name:          "opensuse-tumbleweed",
			want:          tumbleweed,
			wantErr:       false,
			osReleaseFile: "testdata/os-release-opensuse-tumbleweed",
		},
		{
			name:             "alt file",
			want:             fedora,
			wantErr:          false,
			osReleaseFile:    "/does/not/exist",
			osAltReleaseFile: "testdata/os-release-fedora",
		},
		{
			name:             "no files",
			want:             nil,
			wantErr:          true,
			osReleaseFile:    "/does/not/exist",
			osAltReleaseFile: "/also/does/not/exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			OSReleaseFile = tt.osReleaseFile
			OSReleaseAltFile = tt.osAltReleaseFile
			got, err := readOSRelease()

			if (err != nil) != tt.wantErr {
				t.Errorf("readOSRelease() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readOSRelease() = %v, want %v", got, tt.want)
			}
		})
		resetFiles()
	}
}

func TestOSRelease_GetValue(t *testing.T) {
	var fedoraOSRelease, ubuntuOSRelease OSRelease

	var err error

	OSReleaseFile = "testdata/os-release-fedora"

	fedoraOSRelease, err = GetOSRelease()
	require.NoError(t, err)

	OSReleaseFile = "testdata/os-release-ubuntu"

	ubuntuOSRelease, err = GetOSRelease()
	require.NoError(t, err)

	type args struct {
		key string
	}

	tests := []struct {
		name      string
		r         OSRelease
		args      args
		wantValue string
		wantOk    bool
	}{
		{
			name:      "fedora",
			r:         fedoraOSRelease,
			args:      args{key: "NAME"},
			wantValue: "Fedora Linux",
			wantOk:    true,
		},
		{
			name:      "ubuntu",
			r:         ubuntuOSRelease,
			args:      args{key: "NAME"},
			wantValue: "Ubuntu",
			wantOk:    true,
		},
		{
			name:      "unknown key",
			r:         ubuntuOSRelease,
			args:      args{key: "FOOBAR"},
			wantValue: "Unknown",
			wantOk:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.r.GetValue(tt.args.key)
			if gotValue != tt.wantValue {
				t.Errorf("OSRelease.GetValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}

			if gotOk != tt.wantOk {
				t.Errorf("OSRelease.GetValue() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
