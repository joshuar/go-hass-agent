// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMemStats(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		want    memoryStats
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "with swap",
			args: args{file: "testdata/meminfowithswap"},
			want: memoryStats{
				memTotal:     32572792 * 1000,
				memFree:      1396256 * 1000,
				memAvailable: 13353280 * 1000,
				swapTotal:    8388604 * 1000,
				swapCached:   8 * 1000,
				swapFree:     8387836 * 1000,
			},
		},
		{
			name: "without swap",
			args: args{file: "testdata/meminfowithoutswap"},
			want: memoryStats{
				memTotal:     32572792 * 1000,
				memFree:      1396256 * 1000,
				memAvailable: 13353280 * 1000,
			},
		},
		{
			name:    "unavailable",
			args:    args{file: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStatFile = tt.args.file
			got, err := getMemStats(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("getMemStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want[memTotal], got[memTotal])
				assert.Equal(t, tt.want[swapTotal], got[swapTotal])
			}
		})
	}
}
