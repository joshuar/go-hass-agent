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
	wantRequest := &busRequest{
		bus: bus,
	}

	type args struct {
		ctx     context.Context
		busType dbusType
	}
	tests := []struct {
		name string
		args args
		want *busRequest
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

func Test_busRequest_Path(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		p dbus.ObjectPath
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Path(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Match(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		m []dbus.MatchOption
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Match(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Event(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		e string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Event(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Event() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Handler(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		h func(*dbus.Signal)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Handler(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Destination(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		d string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *busRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if got := r.Destination(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.Destination() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Call(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		method string
		args   []any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.Call(tt.args.method, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.Call() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_busRequest_AddWatch(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.AddWatch(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.AddWatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_busRequest_RemoveWatch(t *testing.T) {
	skipCI(t)
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &busRequest{
				bus:          tt.fields.bus,
				eventHandler: tt.fields.eventHandler,
				path:         tt.fields.path,
				event:        tt.fields.event,
				dest:         tt.fields.dest,
				match:        tt.fields.match,
			}
			if err := r.RemoveWatch(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.RemoveWatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetSessionPath(t *testing.T) {
	skipCI(t)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want dbus.ObjectPath
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSessionPath(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSessionPath() = %v, want %v", got, tt.want)
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

// func TestGetData(t *testing.T) {
// 	skipCI(t)
// 	type args struct {
// 		req    *busRequest
// 		method string
// 		args   []any
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    D
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := GetData(tt.args.req, tt.args.method, tt.args.args...)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetData() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetData() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestGetProp(t *testing.T) {
// 	skipCI(t)
// 	type args struct {
// 		req  *busRequest
// 		prop string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    P
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := GetProp(tt.args.req, tt.args.prop)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetProp() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetProp() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
