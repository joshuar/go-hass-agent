// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestNewBusRequest(t *testing.T) {
	skipCI(t)
	ctx := Setup(context.TODO())
	bus, ok := getBus(ctx, SessionBus)
	assert.True(t, ok)
	wantRequest := &BusRequest{
		bus: bus,
	}

	type args struct {
		ctx     context.Context
		busType dbusType
	}
	tests := []struct {
		name string
		args args
		want *BusRequest
	}{
		{
			name: "session request",
			args: args{ctx: ctx, busType: SessionBus},
			want: wantRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBusRequest(tt.args.ctx, tt.args.busType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBusRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariantToValue(t *testing.T) {
	skipCI(t)
	type args struct {
		variant dbus.Variant
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string conversion",
			args: args{variant: dbus.MakeVariant("foo")},
			want: "foo",
		},
		// TODO: Test all possible variant values?
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VariantToValue[string](tt.args.variant); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VariantToValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
