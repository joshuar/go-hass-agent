// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package registry

import (
	"path/filepath"
	"testing"
)

func TestReset(t *testing.T) {
	validPath := t.TempDir()
	newMockReg(t, validPath)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid path",
			args: args{path: validPath},
		},
		{
			name:    "invalid path",
			args:    args{path: filepath.Join(t.TempDir(), "nonexistent")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Reset(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
