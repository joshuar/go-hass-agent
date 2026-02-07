// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package linux

import (
	"testing"
)

func TestFindPortal(t *testing.T) {
	type args struct {
		setup func()
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "KDE",
			args: args{
				setup: func() { t.Setenv("XDG_CURRENT_DESKTOP", "KDE") },
			},
			want: "org.freedesktop.impl.portal.desktop.kde",
		},
		{
			name: "GNOME",
			args: args{
				setup: func() { t.Setenv("XDG_CURRENT_DESKTOP", "GNOME") },
			},
			want: "org.freedesktop.impl.portal.desktop.gtk",
		},
		{
			name: "Unknown",
			args: args{
				setup: func() { t.Setenv("XDG_CURRENT_DESKTOP", "UNKNOWN") },
			},
			want:    "org.freedesktop.impl.portal.desktop.gtk",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()

			ctx := NewContext(t.Context())

			got, err := findPortal(ctx)
			if got != tt.want {
				t.Errorf("FindPortal() = %v, want %v", got, tt.want)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}
