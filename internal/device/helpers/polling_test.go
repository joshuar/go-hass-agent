// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl
//revive:disable:unused-parameter
package helpers

import (
	"context"
	"testing"
	"time"
)

func TestPollSensors(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	updateRan := make(chan struct{})
	update := func(d time.Duration) {
		if d > time.Second {
			close(updateRan)
		}
	}

	type args struct {
		updater  func(time.Duration)
		interval time.Duration
		stdev    time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{updater: update, interval: time.Second, stdev: time.Millisecond},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go PollSensors(ctx, tt.args.updater, tt.args.interval, tt.args.stdev)
			<-updateRan
		})
	}
}
