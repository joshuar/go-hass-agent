// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//nolint:paralleltest
//revive:disable:unused-receiver,comment-spacings
package preferences

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_checkPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exists",
			args: args{path: t.TempDir()},
		},
		{
			name: "does not exist",
			args: args{path: filepath.Join(t.TempDir(), "notexists")},
		},
		{
			name:    "unwriteable",
			args:    args{path: "/notexists"},
			wantErr: true,
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

func TestLoad(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "new file",
			args: args{path: t.TempDir()},
		},
		{
			name: "invalid path",
			args: args{path: "/"},
		},
		{
			name: "existing file",
			args: args{path: "testing/data"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := PathToCtx(context.TODO(), tt.args.path)
			err := Load(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestReset(t *testing.T) {
	existingFilePath := t.TempDir()
	require.NoError(t, checkPath(existingFilePath))
	require.NoError(t,
		os.WriteFile(filepath.Join(existingFilePath, "preferences.toml"),
			[]byte(`existing`),
			0o600))
	nonExistingFilePath := t.TempDir()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "existing file",
			args: args{path: existingFilePath},
		},
		{
			name:    "nonexisting file",
			args:    args{path: nonExistingFilePath},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := PathToCtx(context.TODO(), tt.args.path)
			if err := Reset(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSave(t *testing.T) {
	type args struct {
		path    string
		preload bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "new file",
			args: args{path: t.TempDir(), preload: true},
		},
		{
			name:    "invalid path",
			args:    args{path: "/"},
			wantErr: true,
		},
		{
			name: "existing file",
			args: args{path: "testing/data", preload: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := PathToCtx(context.TODO(), tt.args.path)
			if tt.args.preload {
				require.NoError(t, checkPath(tt.args.path))
				require.NoError(t, Load(ctx))
			}
			if err := Save(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
