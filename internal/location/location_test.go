// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package location

import (
	"context"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

func TestMarshalUpdate(t *testing.T) {
	type args struct {
		l Update
	}
	tests := []struct {
		name string
		args args
		want *hass.LocationUpdate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MarshalUpdate(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendUpdate(t *testing.T) {
	type args struct {
		ctx context.Context
		l   Update
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendUpdate(tt.args.ctx, tt.args.l)
		})
	}
}
