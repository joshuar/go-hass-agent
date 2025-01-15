// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package registry

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

func TestReset(t *testing.T) {
	validPath := t.TempDir()
	ctx := preferences.PathToCtx(context.TODO(), validPath)
	newMockReg(ctx, t)

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
			ctx := preferences.PathToCtx(context.TODO(), tt.args.path)
			if err := Reset(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
