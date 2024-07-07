// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:dupl,paralleltest,wsl
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

func TestGetDistroID(t *testing.T) {
	osReleaseFile := generateMockOSReleaseFile(t)
	tests := []struct {
		name          string
		wantID        string
		wantVersionid string
		osReleaseFile string
	}{
		{
			name:          "File exists",
			wantID:        mockDistroID,
			wantVersionid: mockVersionID,
			osReleaseFile: osReleaseFile,
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

			gotID, gotVersionid := getOSID()
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
	osReleaseFile := generateMockOSReleaseFile(t)

	tests := []struct {
		name          string
		wantName      string
		wantVersion   string
		osReleaseFile string
	}{
		{
			name:          "File exists",
			wantName:      mockDistroName,
			wantVersion:   mockVersion,
			osReleaseFile: osReleaseFile,
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

			gotName, gotVersion := GetOSDetails()
			if gotName != tt.wantName {
				t.Errorf("GetDistroDetails() gotName = %v, want %v", gotName, tt.wantName)
			}

			if gotVersion != tt.wantVersion {
				t.Errorf("GetDistroDetails() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}
