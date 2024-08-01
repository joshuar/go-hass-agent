// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:wsl,paralleltest,nlreturn
package agent

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/scripts"
)

func Test_mergeCh(t *testing.T) {
	ch1 := make(chan int)
	go func() {
		for i := range 5 {
			ch1 <- i
		}
		close(ch1)
	}()

	ch2 := make(chan int)
	go func() {
		for i := range 10 {
			ch2 <- i
		}
		close(ch2)
	}()

	type args struct {
		ctx  context.Context //nolint:containedctx
		inCh []<-chan int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "with input",
			args: args{ctx: context.TODO(), inCh: []<-chan int{ch1, ch2}},
			want: 15,
		},
		{
			name: "without input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			for range mergeCh(tt.args.ctx, tt.args.inCh...) {
				got++
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findScripts(t *testing.T) {
	script, err := scripts.NewScript("testing/data/jsonTestScript.sh")
	require.NoError(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []Script
		wantErr bool
	}{
		{
			name: "with scripts",
			args: args{path: "testing/data"},
			want: []Script{script},
		},
		{
			name: "without scripts",
			args: args{path: "foo/bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findScripts(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("findScripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findScripts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isExecutable(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is executable",
			args: args{filename: "/proc/self/exe"},
			want: true,
		},
		{
			name: "is not executable",
			args: args{filename: "/does/not/exist"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExecutable(tt.args.filename); got != tt.want {
				t.Errorf("isExecutable() = %v, want %v", got, tt.want)
			}
		})
	}
}
