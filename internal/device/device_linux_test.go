// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:dupl,exhaustruct,paralleltest,wsl,nlreturn
package device

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

const (
	mockVersionID  = "9.9"
	mockDistroID   = "testdistro"
	mockVersion    = "9.9 (note)"
	mockDistroName = "Test Distro"
)

func generateMockOSReleaseFile(t *testing.T) string {
	t.Helper()
	mockOSReleaseFile := filepath.Join(t.TempDir(), "goodfile")
	filehandle, err := os.Create(mockOSReleaseFile)
	require.NoError(t, err)
	_, err = fmt.Fprintln(filehandle, `VERSION_ID="`+mockVersionID+`"`)
	require.NoError(t, err)
	_, err = fmt.Fprintln(filehandle, `ID="`+mockDistroID+`"`)
	require.NoError(t, err)
	_, err = fmt.Fprintln(filehandle, `VERSION="`+mockVersion+`"`)
	require.NoError(t, err)
	_, err = fmt.Fprintln(filehandle, `NAME="`+mockDistroName+`"`)
	require.NoError(t, err)

	err = filehandle.Close()
	require.NoError(t, err)

	return mockOSReleaseFile
}

func TestGetOSDetails(t *testing.T) {
	type args struct {
		osReleaseFile string
	}
	tests := []struct {
		name        string
		wantName    string
		wantVersion string
		args        args
		wantErr     bool
	}{
		{
			name:        "File exists",
			wantName:    mockDistroName,
			wantVersion: mockVersion,
			args:        args{osReleaseFile: generateMockOSReleaseFile(t)},
		},
		{
			name:        "File does not exist.",
			wantName:    unknownDistro,
			wantVersion: unknownDistroVersion,
			args:        args{osReleaseFile: "/nonexistent"},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFile := whichdistro.OSReleaseFile
			whichdistro.OSReleaseFile = tt.args.osReleaseFile
			whichdistro.OSReleaseAltFile = tt.args.osReleaseFile
			gotName, gotVersion, err := GetOSDetails()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOSDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("GetOSDetails() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("GetOSDetails() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			whichdistro.OSReleaseFile = origFile
		})
	}
}

func Test_getOSID(t *testing.T) {
	type args struct {
		osReleaseFile string
	}
	tests := []struct {
		name          string
		wantID        string
		wantVersionid string
		args          args
		wantErr       bool
	}{
		{
			name:          "File exists",
			wantID:        mockDistroID,
			wantVersionid: mockVersionID,
			args:          args{osReleaseFile: generateMockOSReleaseFile(t)},
		},
		{
			name:          "File does not exist.",
			wantID:        unknownDistro,
			wantVersionid: unknownDistroVersion,
			args:          args{osReleaseFile: "/nonexistent"},
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFile := whichdistro.OSReleaseFile
			whichdistro.OSReleaseFile = tt.args.osReleaseFile
			whichdistro.OSReleaseAltFile = tt.args.osReleaseFile
			gotID, gotVersionid, err := GetOSID()
			if (err != nil) != tt.wantErr {
				t.Errorf("getOSID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("getOSID() gotId = %v, want %v", gotID, tt.wantID)
			}
			if gotVersionid != tt.wantVersionid {
				t.Errorf("getOSID() gotVersionid = %v, want %v", gotVersionid, tt.wantVersionid)
			}
			whichdistro.OSReleaseFile = origFile
		})
	}
}
