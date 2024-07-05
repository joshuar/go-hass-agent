// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest
package preferences

import (
	_ "embed"
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
			got := DefaultPreferences()
			assert.Equal(t, got.Version, AppVersion)
		})
	}
}
