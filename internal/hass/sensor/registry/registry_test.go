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

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestReset(t *testing.T) {
	appID := "go-hass-agent-test"
	xdg.ConfigHome = t.TempDir()
	ctx := preferences.AppIDToContext(context.TODO(), appID)

	mockReg := newMockReg(ctx, t)

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
			args: args{path: filepath.Dir(mockReg.file)},
		},
		{
			name:    "invalid path",
			args:    args{path: filepath.Join(t.TempDir(), "nonexistent")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Reset(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
