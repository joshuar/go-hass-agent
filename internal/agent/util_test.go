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
