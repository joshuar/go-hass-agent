// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package dbusx

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

//nolint:unused // keep this around
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
		{
			name:    "not string conversion",
			args:    args{variant: dbus.MakeVariant(json.RawMessage(`invalid`))},
			wantErr: true,
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

//revive:disable:function-length
func TestParsePropertiesChanged(t *testing.T) {
	validIntr := "org.some.interface"
	validNewProps := map[string]dbus.Variant{"new": dbus.MakeVariant("value")}
	validOldProps := []string{"old"}

	type args struct {
		propsChanged []any
	}
	tests := []struct {
		wantErrValue error
		want         *Properties
		name         string
		args         args
		wantErr      bool
	}{
		{
			name:         "invalid",
			args:         args{propsChanged: []any{}},
			wantErr:      true,
			wantErrValue: ErrNotPropChanged,
		},
		{
			name:         "bad interface",
			args:         args{propsChanged: []any{-1, validNewProps, validOldProps}},
			wantErr:      true,
			wantErrValue: ErrParseInterface,
		},
		{
			name:         "bad new props",
			args:         args{propsChanged: []any{validIntr, "", validOldProps}},
			wantErr:      true,
			wantErrValue: ErrParseNewProps,
		},
		{
			name:         "bad old props",
			args:         args{propsChanged: []any{validIntr, validNewProps, ""}},
			wantErr:      true,
			wantErrValue: ErrParseOldProps,
		},
		{
			name: "valid",
			args: args{propsChanged: []any{validIntr, validNewProps, validOldProps}},
			want: &Properties{
				Interface:   validIntr,
				Changed:     validNewProps,
				Invalidated: validOldProps,
			},
			wantErr:      false,
			wantErrValue: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePropertiesChanged(tt.args.propsChanged)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePropertiesChanged() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePropertiesChanged() = %v, want %v", got, tt.want)
			}
			assert.ErrorIs(t, err, tt.wantErrValue)
		})
	}
}

func TestParseValueChange(t *testing.T) {
	type args struct {
		valueChanged []any
	}
	tests := []struct {
		wantErrValue error
		want         *Values[string]
		name         string
		args         args
		wantErr      bool
	}{
		{
			name:         "invalid",
			args:         args{valueChanged: []any{}},
			wantErr:      true,
			wantErrValue: ErrNotValChanged,
		},
		{
			name:         "invalid new",
			args:         args{valueChanged: []any{-1, ""}},
			wantErr:      true,
			wantErrValue: ErrParseNewVal,
		},
		{
			name:         "invalid old",
			args:         args{valueChanged: []any{"", -1}},
			wantErr:      true,
			wantErrValue: ErrParseOldVal,
		},
		{
			name:         "valid",
			args:         args{valueChanged: []any{"", ""}},
			want:         &Values[string]{"", ""},
			wantErr:      false,
			wantErrValue: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseValueChange[string](tt.args.valueChanged)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseValueChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseValueChange() = %v, want %v", got, tt.want)
			}
			assert.ErrorIs(t, err, tt.wantErrValue)
		})
	}
}
