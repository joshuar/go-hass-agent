// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wsl,unused,nlreturn
package dbusx

import (
	"os"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func skipCI(t *testing.T) {
	t.Helper()

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestVariantToValue(t *testing.T) {
	type args struct {
		variant dbus.Variant
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "string conversion",
			args: args{variant: dbus.MakeVariant("foo")},
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VariantToValue[string](tt.args.variant)
			if (err != nil) != tt.wantErr {
				t.Errorf("VariantToValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VariantToValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
