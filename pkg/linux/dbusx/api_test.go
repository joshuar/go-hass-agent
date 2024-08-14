// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package dbusx

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDBusAPI(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger *slog.Logger
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "successful test",
			args: args{ctx: context.TODO(), logger: slog.Default()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDBusAPI(tt.args.ctx, tt.args.logger)
			for _, b := range []dbusType{SessionBus, SystemBus} {
				bus, err := got.GetBus(context.TODO(), b)
				require.NoError(t, err)
				assert.Equal(t, b, bus.busType)
				assert.True(t, bus.conn.Connected())
			}
		})
	}
}

func TestDBusAPI_GetBus(t *testing.T) {
	validAPI := NewDBusAPI(context.TODO(), slog.Default())
	emptyAPI := &DBusAPI{dbus: map[dbusType]*Bus{}}

	type fields struct {
		dbus map[dbusType]*Bus
	}
	type args struct {
		ctx     context.Context
		busType dbusType
	}
	tests := []struct {
		fields  fields
		args    args
		name    string
		wantErr bool
	}{
		{
			name:   "valid api",
			args:   args{ctx: context.TODO(), busType: SessionBus},
			fields: fields{dbus: validAPI.dbus},
		},
		{
			name:   "empty api",
			args:   args{ctx: context.TODO(), busType: SessionBus},
			fields: fields{dbus: emptyAPI.dbus},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DBusAPI{
				dbus: tt.fields.dbus,
			}
			got, err := a.GetBus(tt.args.ctx, tt.args.busType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBusAPI.GetBus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.args.busType, got.busType)
			assert.True(t, got.conn.Connected())
		})
	}
}
