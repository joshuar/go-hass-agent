// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func Test_newBus(t *testing.T) {
	type args struct {
		ctx context.Context
		t   dbusType
	}
	tests := []struct {
		name string
		args args
		want *Bus
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBus(tt.args.ctx, tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newBus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBusRequest(t *testing.T) {
	type args struct {
		ctx     context.Context
		busType dbusType
	}
	tests := []struct {
		name string
		args args
		want *busRequest
	}{
		// TODO: Add test cases.
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

func Test_busRequest_GetProp(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		prop string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    dbus.Variant
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
			got, err := r.GetProp(tt.args.prop)
			if (err != nil) != tt.wantErr {
				t.Errorf("busRequest.GetProp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.GetProp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_SetProp(t *testing.T) {
	type fields struct {
		bus          *Bus
		eventHandler func(*dbus.Signal)
		path         dbus.ObjectPath
		event        string
		dest         string
		match        []dbus.MatchOption
	}
	type args struct {
		prop  string
		value dbus.Variant
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
			if err := r.SetProp(tt.args.prop, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("busRequest.SetProp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_busRequest_GetData(t *testing.T) {
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
		args   []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dbusData
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
			if got := r.GetData(tt.args.method, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busRequest.GetData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_busRequest_Call(t *testing.T) {
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
		args   []interface{}
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

func Test_dbusData_AsVariantMap(t *testing.T) {
	type fields struct {
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]dbus.Variant
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsVariantMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsVariantMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsStringMap(t *testing.T) {
	type fields struct {
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsStringMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsStringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsObjectPathList(t *testing.T) {
	type fields struct {
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   []dbus.ObjectPath
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsObjectPathList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsObjectPathList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsStringList(t *testing.T) {
	type fields struct {
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsStringList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbusData_AsObjectPath(t *testing.T) {
	type fields struct {
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   dbus.ObjectPath
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dbusData{
				data: tt.fields.data,
			}
			if got := d.AsObjectPath(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbusData.AsObjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_variantToValue(t *testing.T) {
// 	type args struct {
// 		variant dbus.Variant
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want S
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := variantToValue(tt.args.variant); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("variantToValue() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_findPortal(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findPortal(); got != tt.want {
				t.Errorf("findPortal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHostname(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHostname(tt.args.ctx); got != tt.want {
				t.Errorf("GetHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHardwareDetails(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetHardwareDetails(tt.args.ctx)
			if got != tt.want {
				t.Errorf("GetHardwareDetails() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetHardwareDetails() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
