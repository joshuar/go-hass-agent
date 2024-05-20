// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"os"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
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
