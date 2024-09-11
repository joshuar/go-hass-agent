// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type args struct {
		id      string
		options Options
	}
	tests := []struct {
		want *slog.Logger
		name string
		args args
	}{
		{
			name: "with log file",
			args: args{id: "go-hass-agent-test"},
		},
		{
			name: "with log file and custom level",
			args: args{id: "go-hass-agent-test", options: Options{LogLevel: "debug", NoLogFile: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.id, tt.args.options)
			switch tt.args.options.LogLevel {
			case "debug":
				got.Debug("Test Message")
				slog.Debug("Via default")
			default:
				got.Info("Test Message")
				slog.Info("Via default")
			}
			if tt.args.options.NoLogFile {
				return
			}
			data, err := os.ReadFile(filepath.Join(xdg.ConfigHome, tt.args.id, "agent.log"))
			require.NoError(t, err)
			assert.Contains(t, string(data), string("Test Message"))
			assert.Contains(t, string(data), string("Via default"))
		})
	}
}

func Test_openLogFile(t *testing.T) {
	type args struct {
		logFile string
	}
	tests := []struct {
		want    *os.File
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "unwriteable directory",
			args:    args{logFile: "/sys/test.log"},
			wantErr: true,
		},
		{
			name:    "unwriteable file",
			args:    args{logFile: "/sys/device/test.log"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openLogFile(tt.args.logFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("openLogFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("openLogFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset(t *testing.T) {
	deleteableFile := filepath.Join(t.TempDir(), "test.log")
	err := os.WriteFile(deleteableFile, []byte(`test`), 0o600)
	require.NoError(t, err)

	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "with deleteable log file",
			args: args{file: deleteableFile},
		},
		{
			name: "without log file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Reset(tt.args.file); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.file != "" {
				assert.NoFileExists(t, tt.args.file)
			}
		})
	}
}
